package metricscli

import (
	"log"
	"time"

	"github.com/AgentGuo/scheduler/cmd/metrics-cli/config"
	"github.com/AgentGuo/scheduler/util"
	"github.com/go-redis/redis"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
	"k8s.io/apimachinery/pkg/util/json"
)

const (
	MetricsInfoKey = "metricsInfo"
	NodeInfoKey    = "nodeInfo"
	TaskInfoKey    = "taskInfo"
)

type MetricsInfo struct {
	CpuRemain int64 `json:"cpu_remain"`
	MemFree   int64 `json:"mem_free"`
	TimeStamp int64 `json:"time_stamp"`
}

type NodeInfo struct {
	HostIP    string `json:"host_ip"`
	Cpu       int64  `json:"cpu"`
	Mem       int64  `json:"mem"`
	TimeStamp int64  `json:"time_stamp"`
}

var (
	coreNums int
	memTotal int64
	hostName string
)

// 初始化节点信息
func initMetricsCli() error {
	hostInfo, err := host.Info()
	if err != nil {
		log.Fatal(err)
		return err
	}
	hostName = hostInfo.Hostname
	coreNums, err = cpu.Counts(true)
	if err != nil {
		log.Fatal(err)
		return err
	} else {
		coreNums *= 1000 // 单位为mCPU
	}
	memStat, err := mem.VirtualMemory()
	if err != nil {
		log.Fatal(err)
		return err
	}
	// memTotal = float64(memStat.Total) / (1024 * 1024)
	memTotal = int64(memStat.Total) // 单位为byte
	return nil
}

func RunMetricsCli(config *config.MetricsCliConfig) {
	cli := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
		Password: config.RedisPassword,
	})
	pong, err := cli.Ping().Result()
	if err != nil {
		log.Fatal(err)
		return
	} else {
		log.Println(pong)
	}
	err = initMetricsCli()
	if err != nil {
		return
	}
	// 提交节点信息
	err = emitNodeInfo(cli)
	if err != nil {
		return
	}
	// 定时提交监控信息
	for {
		emitMetricsInfo(cli)
	}
}

func emitNodeInfo(cli *redis.Client) error {
	ip, err := util.GetLocalIP()
	if err != nil {
		log.Fatal(err)
	}
	v, _ := json.Marshal(NodeInfo{
		HostIP:    ip,
		Cpu:       int64(coreNums),
		Mem:       memTotal,
		TimeStamp: time.Now().Unix(),
	})
	err = cli.HSet(NodeInfoKey, hostName, v).Err()
	if err != nil {
		log.Fatal(err)
	}
	return err
}

func emitMetricsInfo(cli *redis.Client) error {
	var (
		cpuRemain float64 = 0
		memFree   int64   = 0
	)
	go func() {
		totalPercent, err := cpu.Percent(time.Second, false)
		if err != nil {
			log.Fatal(err)
			return
		}
		cpuRemain = float64(coreNums) * (100 - totalPercent[0]) / 100
	}()
	go func() {
		memStat, err := mem.VirtualMemory()
		if err != nil {
			log.Fatal()
			return
		}
		// memFree = float64(memStat.Free) / (1024 * 1024)
		memFree = int64(memStat.Free)
	}()
	time.Sleep(2 * time.Second)
	// log.Printf("%s: cpu remaining amount(mCPU): %.2f m, mem free: %.2f MB\n", hostName, cpuRemain, memFree)
	log.Printf("%s: cpu remaining amount(mCPU): %.2f m, mem free: %d Bytes\n", hostName, cpuRemain, memFree)
	v, _ := json.Marshal(MetricsInfo{
		CpuRemain: int64(cpuRemain),
		MemFree:   memFree,
		TimeStamp: time.Now().Unix(),
	})
	err := cli.HSet(MetricsInfoKey, hostName, v).Err()
	if err != nil {
		log.Fatal(err)
	}
	return err
}

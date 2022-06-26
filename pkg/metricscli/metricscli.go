package metricscli

import (
	"github.com/AgentGuo/scheduler/cmd/metrics-cli/config"
	"github.com/go-redis/redis"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
	"k8s.io/apimachinery/pkg/util/json"
	"log"
	"time"
)

type MetricsInfo struct {
	CpuRemain float64 `json:"cpu_remain"`
	MemFree   float64 `json:"mem_free"`
	TimeStamp int64   `json:"time_stamp"`
}

func RunMetricsCli(config *config.MetricsCliConfig) {
	cli := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
		Password: config.RedisPassword,
	})
	pong, err := cli.Ping().Result()
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println(pong)
	}
	hostInfo, err := host.Info()
	if err != nil {
		log.Fatal(err)
		return
	}
	hostName := hostInfo.Hostname
	coreNums, err := cpu.Counts(true)
	if err != nil {
		log.Fatal(err)
		return
	} else {
		coreNums *= 1000 // 单位为mCPU
	}
	var (
		cpuRemain float64 = 0
		memFree   float64 = 0
	)
	for {
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
			memFree = float64(memStat.Free) / (1024 * 1024)
		}()
		time.Sleep(2 * time.Second)
		log.Printf("%s: cpu remaining amount(mCPU): %.2f m, mem free: %.2f MB\n", hostName, cpuRemain, memFree)
		v, _ := json.Marshal(MetricsInfo{
			CpuRemain: cpuRemain,
			MemFree:   memFree,
			TimeStamp: time.Now().Unix(),
		})
		err = cli.Set(hostName, string(v), 0).Err()
		if err != nil {
			log.Fatal(err)
		}
	}
}

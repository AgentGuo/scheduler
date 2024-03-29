package metricscli

import (
	"context"
	"github.com/AgentGuo/scheduler/cmd/metrics-cli/config"
	"github.com/AgentGuo/scheduler/util"
	"github.com/go-redis/redis"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"path/filepath"
	"time"
)

const (
	MetricsInfoKey = "metricsInfo"
	NodeInfoKey    = "nodeInfo"
)

type MetricsInfo struct {
	CpuRemain float64           `json:"cpu_remain"`
	MemFree   float64           `json:"mem_free"`
	Labels    map[string]string `json:"labels"`
	TimeStamp int64             `json:"time_stamp"`
}

type NodeInfo struct {
	Cpu       float64 `json:"cpu"`
	Mem       float64 `json:"mem"`
	TimeStamp int64   `json:"time_stamp"`
}

var (
	coreNums int
	memTotal float64
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
	memTotal = float64(memStat.Total) / (1024 * 1024)
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
	v, _ := json.Marshal(NodeInfo{
		Cpu:       float64(coreNums),
		Mem:       memTotal,
		TimeStamp: time.Now().Unix(),
	})
	err := cli.HSet(NodeInfoKey, hostName, v).Err()
	if err != nil {
		log.Fatal(err)
	}
	return err
}

func emitMetricsInfo(cli *redis.Client) error {
	var (
		cpuRemain float64 = 0
		memFree   float64 = 0
		labels    map[string]string
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
		memFree = float64(memStat.Free) / (1024 * 1024)
	}()
	go func() {
		kubeCli := newKubeCli()
		node, err := kubeCli.CoreV1().Nodes().Get(context.Background(), hostName, metav1.GetOptions{})
		if err != nil {
			log.Fatal(err)
		}
		labels = node.Labels
	}()
	time.Sleep(2 * time.Second)
	log.Printf("%s: cpu remaining amount(mCPU): %.2f m, mem free: %.2f MB\n", hostName, cpuRemain, memFree)
	v, _ := json.Marshal(MetricsInfo{
		CpuRemain: cpuRemain,
		MemFree:   memFree,
		Labels:    labels,
		TimeStamp: time.Now().Unix(),
	})
	err := cli.HSet(MetricsInfoKey, hostName, v).Err()
	if err != nil {
		log.Fatal(err)
	}
	return err
}

func newKubeCli() *kubernetes.Clientset {
	kubeConfig := filepath.Join(util.HomeDir(), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err != nil {
		panic(err.Error())
	}
	client, err := kubernetes.NewForConfig(config)
	return client
}

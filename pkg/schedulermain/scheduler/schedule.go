package scheduler

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/AgentGuo/scheduler/cmd/scheduler-main/config"
	"github.com/AgentGuo/scheduler/pkg/metricscli"
	"github.com/AgentGuo/scheduler/pkg/resourcemanage/apis"
	"github.com/AgentGuo/scheduler/pkg/schedulermain/resourcemanagecli"
	"github.com/AgentGuo/scheduler/pkg/schedulermain/scheduler/plugin"
	"github.com/AgentGuo/scheduler/pkg/schedulermain/task"
	"github.com/AgentGuo/scheduler/pkg/schedulermain/task/kubequeue"
	"github.com/go-redis/redis"
	"k8s.io/apimachinery/pkg/util/json"
)

type Scheduler struct {
	RedisCli       *redis.Client
	ScorePlugins   []plugin.ScorePlugin
	ScoreWeights   []float64
	FilterPlugins  []plugin.FilterPlugin
	ResourceClient *resourcemanagecli.ResourceClient
}

func NewScheduler(cfg *config.SchedulerMainConfig) (*Scheduler, error) {
	redisCli := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
	})
	_, err := redisCli.Ping().Result()
	if err != nil {
		return nil, err
	}
	pluginRegMap := GetRegistryMap()
	scorePlugins, scoreWeights := InitScorePlugin(redisCli, cfg, pluginRegMap)
	filterPlugins := InitFilterPlugin(redisCli, cfg, pluginRegMap)
	return &Scheduler{
		RedisCli:       redisCli,
		ScorePlugins:   scorePlugins,
		ScoreWeights:   scoreWeights,
		FilterPlugins:  filterPlugins,
		ResourceClient: resourcemanagecli.NewResourceClient(cfg.ResourceManagerPort),
	}, nil
}

func (s *Scheduler) Schedule(ctx context.Context, t *task.Task) (nodeName string, err error) {
	if t.TaskType == task.NormalTaskType {
		nodeList := ListLiveNode(s.RedisCli)
		nodeList = s.filter(ctx, nodeList, t)
		nodeList, err = s.score(ctx, nodeList, t)
		if err != nil {
			return "", err
		}
		if len(nodeList) == 0 {
			return "", fmt.Errorf("no node for scheduling")
		}
		return nodeList[0], nil
	} else if t.TaskType == task.KubeResourceTaskType {
		// hostName := s.FindHostNameByTaskName(t.Name)
		// if hostName == "" {
		// 	return "", fmt.Errorf("not found hostName by taskName{%s}", t.Name)
		// }
		// return hostName, nil
		return "", nil
	}

	return "", fmt.Errorf("wrong task type")
}

func ListLiveNode(client *redis.Client) []string {
	res := []string{}
	keys, err := client.HKeys(metricscli.MetricsInfoKey).Result()
	if err != nil {
		return res
	}
	for _, key := range keys {
		v, err := client.HGet(metricscli.MetricsInfoKey, key).Result()
		if err != nil {
			continue
		}
		nodeInfo := &metricscli.MetricsInfo{}
		err = json.Unmarshal([]byte(v), nodeInfo)
		if err != nil {
			continue
		}
		// 说明主机在线
		if time.Now().Add(-time.Second*5).Unix() < nodeInfo.TimeStamp {
			res = append(res, key)
		}
	}
	return res
}

func (s *Scheduler) ExecuteResourceT(ctx context.Context, t *task.Task) error {
	// logger, _ := util.GetCtxLogger(ctx)
	taskInfo := s.FindTaskInfoByTaskName(t.Name)
	if taskInfo == nil {
		return fmt.Errorf("not found TaskInfo by taskName{%s}", t.Name)
	}
	// check task
	if taskInfo.Status != task.RUNNING {
		return fmt.Errorf("task %+v is not running", taskInfo)
	}

	// get node static information
	nodeInfo := s.FindNodeInfoByHostName(t.NodeName)
	if nodeInfo == nil {
		return fmt.Errorf("not found NodeInfo by hostName{%s}", t.NodeName)
	}

	// get node dynamic information
	metricsInfo := s.FindMetricsInfoByHostName(t.NodeName)
	if metricsInfo == nil {
		return fmt.Errorf("not found MetricsInfo by hostName{%s}", t.NodeName)
	}

	// get old resource limit
	resourceDetail, ok := taskInfo.ResourceDetail.(apis.ResourceValue)
	if !ok {
		return fmt.Errorf("taskInfo.ResourceDetail can not convert to ResourceValue")
	}

	checkInfo := &CheckReourceInfo{
		OldCpuLimit: resourceDetail.CpuLimit,
		OldMemLimit: resourceDetail.MemoryLimit,
	}

	// kubernetes type
	if t.TaskType == task.KubeResourceTaskType {
		kubeDetail, ok := t.Detail.(apis.KubeResourceTask)
		if !ok {
			return fmt.Errorf("t.Detail can not convert to KubeResourceTask")
		}

		checkInfo.NewCpuLimit = kubeDetail.CpuLimit
		checkInfo.NewMemLimit = kubeDetail.MemoryLimit
		// check resource
		err := s.checkReource(ctx, nodeInfo, metricsInfo, checkInfo)
		if err != nil {
			return err
		}
	} else {
		// TODO
	}

	err := s.ResourceClient.Execute(ctx, t, nodeInfo.HostIP)
	if err != nil {
		return err
	}

	// excute success, update database
	taskInfo.ResourceDetail = apis.ResourceValue{
		CpuLimit:    checkInfo.NewCpuLimit,
		MemoryLimit: checkInfo.NewMemLimit,
	}
	taskInfo.UpdateTime = time.Now().Unix()

	// update database
	v, err := json.Marshal(*taskInfo)
	if err != nil {
		return err
	}
	err = s.RedisCli.HSet(metricscli.TaskInfoKey, taskInfo.Name, v).Err()
	if err != nil {
		return err
	}
	return nil
}

func (s *Scheduler) FindTaskInfoByTaskName(taskName string) *task.Task {
	v, err := s.RedisCli.HGet(metricscli.TaskInfoKey, taskName).Result()
	if err != nil || v == "" {
		return nil
	}
	taskInfo := &task.Task{}
	err = json.Unmarshal([]byte(v), taskInfo)
	if err != nil {
		log.Printf("unmarshal TaskInfo failed\n")
		return nil
	}
	// 注意: 此时taskInfo中的Detail和ResourceDetail是map[string]interface类型, 且CpuLimit和MemoryLimit都是float64
	tm, ok := taskInfo.ResourceDetail.(map[string]interface{})
	if !ok {
		log.Println("taskInfo.ResourceDetail can not convert to map[string]interface{}")
		return nil
	}
	taskInfo.ResourceDetail = apis.ResourceValue{
		CpuLimit:    int64(tm["CpuLimit"].(float64)),
		MemoryLimit: int64(tm["MemoryLimit"].(float64)),
	}
	tm, ok = taskInfo.Detail.(map[string]interface{})
	if !ok {
		log.Println("taskInfo.Detail can not convert to map[string]interface{}")
		return nil
	}
	taskInfo.Detail = kubequeue.KubeTaskDetails{
		PodName:   tm["PodName"].(string),
		UID:       tm["UID"].(string),
		Namespace: tm["Namespace"].(string),
	}
	return taskInfo
}

func (s *Scheduler) FindNodeInfoByHostName(hostName string) *metricscli.NodeInfo {
	v, err := s.RedisCli.HGet(metricscli.NodeInfoKey, hostName).Result()
	if err != nil || v == "" {
		return nil
	}
	hostInfo := &metricscli.NodeInfo{}
	err = json.Unmarshal([]byte(v), hostInfo)
	if err != nil {
		log.Printf("unmarshal NodeInfo failed\n")
		return nil
	}
	return hostInfo
}

func (s *Scheduler) FindMetricsInfoByHostName(hostName string) *metricscli.MetricsInfo {
	v, err := s.RedisCli.HGet(metricscli.MetricsInfoKey, hostName).Result()
	if err != nil || v == "" {
		return nil
	}
	metricsInfo := &metricscli.MetricsInfo{}
	err = json.Unmarshal([]byte(v), metricsInfo)
	if err != nil {
		log.Printf("unmarshal NodeInfo failed\n")
		return nil
	}
	return metricsInfo
}

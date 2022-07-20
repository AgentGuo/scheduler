package schedule

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/AgentGuo/scheduler/cmd/scheduler-main/config"
	"github.com/AgentGuo/scheduler/pkg/metricscli"
	"github.com/AgentGuo/scheduler/pkg/resourcemanage/apis"
	"github.com/AgentGuo/scheduler/pkg/schedulermain/resourcemanagecli"
	"github.com/AgentGuo/scheduler/task"
	"github.com/AgentGuo/scheduler/task/kubequeue"
	"github.com/go-redis/redis"
)

type Scheduler struct {
	RedisCli       *redis.Client
	ResourceClient *resourcemanagecli.ResourceClient
}

func NewScheduler(config *config.SchedulerMainConfig) (*Scheduler, error) {
	redisCli := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
		Password: config.RedisPassword,
	})
	_, err := redisCli.Ping().Result()
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	return &Scheduler{
		RedisCli:       redisCli,
		ResourceClient: resourcemanagecli.NewResourceClient(config.ResourceManagerPort),
	}, nil
}

func (s *Scheduler) Schedule(t *task.Task) (string, error) {
	if t.TaskType == task.NormalTaskType {
		priorityList, err := s.score()
		if err != nil {
			return "", err
		}
		if len(priorityList) == 0 {
			return "", fmt.Errorf("no node for scheduling")
		}
		return priorityList[0].NodeName, nil
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

func (s *Scheduler) ExecuteResourceT(t *task.Task) error {
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
		err := s.checkReource(nodeInfo, metricsInfo, checkInfo)
		if err != nil {
			return err
		}
	} else {
		// TODO
	}

	err := s.ResourceClient.Execute(t, nodeInfo.HostIP)
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

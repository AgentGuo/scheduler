package schedulermain

import (
	"context"
	"fmt"
	"github.com/AgentGuo/scheduler/cmd/scheduler-main/config"
	"github.com/AgentGuo/scheduler/pkg/metricscli"
	"github.com/AgentGuo/scheduler/pkg/schedulermain/kubebinder"
	"github.com/AgentGuo/scheduler/pkg/schedulermain/scheduler"
	"github.com/AgentGuo/scheduler/pkg/schedulermain/task"
	"github.com/AgentGuo/scheduler/pkg/schedulermain/task/kubequeue"
	"github.com/AgentGuo/scheduler/pkg/schedulermain/task/queue"
	"github.com/AgentGuo/scheduler/pkg/schedulermain/task/resourcequeue"
	"github.com/AgentGuo/scheduler/util"
	"k8s.io/apimachinery/pkg/util/json"
	"time"
)

type SchedulerMain struct {
	ScheduleQueueList queue.ScheduleQueueList
	Binder            Binder
	Scheduler         *scheduler.Scheduler
}

func NewSchedulerMain(ctx context.Context, config *config.SchedulerMainConfig) (*SchedulerMain, error) {
	// step1: 初始化scheduler
	Scheduler, err := scheduler.NewScheduler(config)
	if err != nil {
		return nil, err
	}
	// step2: 初始化调度队列
	scheduleQueueList := queue.ScheduleQueueList{}
	err = resourcequeue.AppendResourceQueue(ctx, &scheduleQueueList, config.ResourceQueueServerPort)
	if err != nil {
		return nil, err
	}
	err = kubequeue.AppendKubeQueue(ctx, &scheduleQueueList)
	if err != nil {
		return nil, err
	}
	// step3: 初始化binder
	binder, err := kubebinder.NewKubeBind()
	if err != nil {
		return nil, err
	}
	return &SchedulerMain{
		ScheduleQueueList: scheduleQueueList,
		Binder:            binder,
		Scheduler:         Scheduler,
	}, nil
}

func RunSchedulerMain(ctx context.Context, config *config.SchedulerMainConfig) {
	fmt.Printf("hello, i am scheduler main! this is the config file\n%+v\n", config)
	logger, _ := util.GetCtxLogger(ctx)
	s, err := NewSchedulerMain(ctx, config)
	if err != nil {
		logger.Fatalf("new scheduler failed, err:%+v", err)
		return
	}
	for {
		// step0: 从多个调度队列中取出调度任务
		t, err := s.ScheduleQueueList.GetListTask()
		if err != nil {
			logger.Errorf("get task from queue list error: %+v", err)
			continue
		}
		scheLogger, _ := util.GetCtxLogger(ctx)
		switch t.TaskType {
		case task.NormalTaskType:
			// step1: 调度到一个节点
			util.SetCtxFields(ctx, map[string]string{task.TaskNameLogKey: t.Name})
			node, err := s.Scheduler.Schedule(ctx, t)
			if err != nil {
				scheLogger.Errorf("schedule failed:%+v", err)
				continue
			}
			// step2: 绑定到节点
			err = s.Binder.Bind(t, node)
			if err != nil {
				scheLogger.Errorf("bind failed:%+v", err)
				continue
			}
			t.NodeName = node
			t.Status = task.RUNNING
			t.UpdateTime = time.Now().Unix()
			v, err := json.Marshal(*t)
			if err != nil {
				// 失败处理
				scheLogger.Errorf("json marshal failed:%+v", err)
				continue
			}
			err = s.Scheduler.RedisCli.HSet(metricscli.TaskInfoKey, t.Name, v).Err()
			if err != nil {
				scheLogger.Errorf("update taskinfo failed:%+v", err)
			}
		case task.KubeResourceTaskType:
			_, err := s.Scheduler.Schedule(ctx, t)
			if err != nil {
				scheLogger.Errorf("schedule failed:%+v", err)
				continue
			}
			err = s.Scheduler.ExecuteResourceT(ctx, t)
			if err != nil {
				scheLogger.Errorf("execute resource task failed:%+v", err)
			}
		}
	}
}

package schedulermain

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/AgentGuo/scheduler/cmd/scheduler-main/config"
	"github.com/AgentGuo/scheduler/pkg/metricscli"
	"github.com/AgentGuo/scheduler/pkg/schedulermain/kubebinder"
	"github.com/AgentGuo/scheduler/pkg/schedulermain/scheduler"
	"github.com/AgentGuo/scheduler/pkg/schedulermain/task"
	"github.com/AgentGuo/scheduler/pkg/schedulermain/task/kubequeue"
	"github.com/AgentGuo/scheduler/pkg/schedulermain/task/queue"
	"github.com/AgentGuo/scheduler/pkg/schedulermain/task/resourcequeue"
	"github.com/AgentGuo/scheduler/util"
)

type SchedulerMain struct {
	ScheduleQueueList queue.ScheduleQueueList
	Binder            Binder
	Scheduler         *scheduler.Scheduler
}

func NewSchedulerMain(ctx context.Context, config *config.SchedulerMainConfig) (*SchedulerMain, error) {
	Scheduler, err := scheduler.NewScheduler(config)
	if err != nil {
		return nil, err
	}
	rq, err := resourcequeue.NewResourceQueue(ctx, config.ResourceQueueServerPort)
	if err != nil {
		return nil, err
	}
	kq, err := kubequeue.NewKubeQueue(ctx)
	if err != nil {
		return nil, err
	}
	return &SchedulerMain{
		ScheduleQueueList: []queue.ScheduleQueue{rq, kq},
		Binder:            kubebinder.NewKubeBind(),
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
		t := s.ScheduleQueueList.GetListTask()
		if t != nil {
			scheLogger, _ := util.GetCtxLogger(ctx)
			if t.TaskType == task.NormalTaskType {
				util.SetCtxFields(ctx, map[string]string{task.TaskNameLogKey: t.Name})
				if t.Name == "pause-default" { // 测试用的逻辑
					node, err := s.Scheduler.Schedule(ctx, t)
					if err != nil {
						scheLogger.Errorf("schedule failed:%+v", err)
						continue
					}
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
				}
			} else if t.TaskType == task.KubeResourceTaskType {
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
		} else {
			time.Sleep(time.Second)
		}
	}
}

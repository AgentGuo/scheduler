package schedulermain

import (
	"context"
	"fmt"
	"github.com/AgentGuo/scheduler/cmd/scheduler-main/config"
	"github.com/AgentGuo/scheduler/pkg/schedulermain/kubebinder"
	"github.com/AgentGuo/scheduler/pkg/schedulermain/scheduler"
	"github.com/AgentGuo/scheduler/pkg/schedulermain/task"
	"github.com/AgentGuo/scheduler/pkg/schedulermain/task/kubequeue"
	"github.com/AgentGuo/scheduler/pkg/schedulermain/task/queue"
	"github.com/AgentGuo/scheduler/util"
	"time"
)

type SchedulerMain struct {
	ScheduleQ queue.ScheduleQueue
	Binder    Binder
	Scheduler *scheduler.Scheduler
}

const taskInfoKey = "taskInfo"

func NewSchedulerMain(ctx context.Context, config *config.SchedulerMainConfig) (*SchedulerMain, error) {
	Scheduler, err := scheduler.NewScheduler(config)
	if err != nil {
		return nil, err
	}
	return &SchedulerMain{
		ScheduleQ: kubequeue.NewKubeQueue(ctx),
		Binder:    kubebinder.NewKubeBind(),
		Scheduler: Scheduler,
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
		t := s.ScheduleQ.GetTask()
		if t != nil {
			util.SetCtxFields(ctx, map[string]string{task.TaskNameLogKey: t.Name})
			scheLogger, _ := util.GetCtxLogger(ctx)
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
				err = s.Scheduler.RedisCli.HSet(taskInfoKey, t.Name, node).Err()
				if err != nil {
					scheLogger.Errorf("task schedule result write failed:%+v", err)
				}
			}
		}
		time.Sleep(time.Second)
	}
}

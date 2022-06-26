package schedulermain

import (
	"github.com/AgentGuo/scheduler/cmd/scheduler-main/config"
	"github.com/AgentGuo/scheduler/pkg/schedulermain/binder"
	"github.com/AgentGuo/scheduler/pkg/schedulermain/kubebinder"
	"github.com/AgentGuo/scheduler/pkg/schedulermain/schedule"
	"github.com/AgentGuo/scheduler/task/kubequeue"
	"github.com/AgentGuo/scheduler/task/queue"
	"log"
	"time"
)

type SchedulerMain struct {
	ScheduleQ queue.ScheduleQueue
	Binder    binder.Binder
	Scheduler *schedule.Scheduler
}

func NewSchedulerMain(config *config.SchedulerMainConfig) (*SchedulerMain, error) {
	Scheduler, err := schedule.NewScheduler(config)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	return &SchedulerMain{
		ScheduleQ: kubequeue.NewKubeQueue(),
		Binder:    kubebinder.NewKubeBind(),
		Scheduler: Scheduler,
	}, nil
}

func RunSchedulerMain(config *config.SchedulerMainConfig) {
	log.Println("hello, i am scheduler main! this is the config file")
	log.Printf("%+v\n", config)
	s, err := NewSchedulerMain(config)
	if err != nil {
		log.Fatal(err)
		return
	}
	for {
		t := s.ScheduleQ.GetTask()
		if t != nil {
			log.Printf("unscheduled pod: %s\n", t.Name)
			if t.Name == "pause" {
				node, err := s.Scheduler.Schedule(t)
				if err != nil {
					log.Fatal(err)
					continue
				}
				err = s.Binder.Bind(t, node)
				if err != nil {
					log.Fatal(err)
					continue
				}
			}
		}
		time.Sleep(time.Second)
	}
}

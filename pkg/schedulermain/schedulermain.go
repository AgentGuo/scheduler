package schedulermain

import (
	"encoding/json"
	"log"
	"time"

	"github.com/AgentGuo/scheduler/cmd/scheduler-main/config"
	"github.com/AgentGuo/scheduler/pkg/metricscli"
	"github.com/AgentGuo/scheduler/pkg/schedulermain/binder"
	"github.com/AgentGuo/scheduler/pkg/schedulermain/kubebinder"
	"github.com/AgentGuo/scheduler/pkg/schedulermain/schedule"
	"github.com/AgentGuo/scheduler/task"
	"github.com/AgentGuo/scheduler/task/kubequeue"
	"github.com/AgentGuo/scheduler/task/queue"
)

type SchedulerMain struct {
	ScheduleQ queue.ScheduleQueue
	Binder    binder.Binder
	Scheduler *schedule.Scheduler
}

func NewSchedulerMain(config *config.SchedulerMainConfig) (*SchedulerMain, error) {
	log.Println("hello, i am scheduler main! this is the config file")
	log.Printf("%+v\n", config)
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

func (s *SchedulerMain) Run() {
	for {
		t := s.ScheduleQ.GetTask()

		if t != nil {
			if t.TaskType == task.NormalTaskType {
				log.Printf("unscheduled task: %s\n", t.Name)
				if t.Name == "pause-default" { // 测试用的逻辑
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
					t.NodeName = node
					t.Status = task.RUNNING
					t.UpdateTime = time.Now().Unix()
					v, err := json.Marshal(*t)
					if err != nil {
						// 失败处理
						log.Fatal(err)
						continue
					}
					err = s.Scheduler.RedisCli.HSet(metricscli.TaskInfoKey, t.Name, v).Err()
					if err != nil {
						log.Fatal(err)
					}
				}
			} else if t.TaskType == task.KubeResourceTaskType {
				_, err := s.Scheduler.Schedule(t)
				if err != nil {
					log.Fatal(err)
					continue
				}
				err = s.Scheduler.ExecuteResourceT(t)
				if err != nil {
					log.Fatal(err)
				}
			}
		}
		time.Sleep(time.Second)
	}
}

func (s *SchedulerMain) SubmitTask(task task.Task) {
	s.ScheduleQ.SubmitTask(task)
}

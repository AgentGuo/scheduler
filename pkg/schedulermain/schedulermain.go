package schedulermain

import (
	"github.com/AgentGuo/scheduler/cmd/scheduler-main/config"
	"github.com/AgentGuo/scheduler/pkg/schedulermain/binder"
	"github.com/AgentGuo/scheduler/pkg/schedulermain/kubebinder"
	"github.com/AgentGuo/scheduler/task/kubequeue"
	"github.com/AgentGuo/scheduler/task/queue"
	"log"
)

func RunSchedulerMain(config *config.SchedulerMainConfig) {
	log.Println("hello, i am binder main! this is the config file")
	log.Printf("%v\n", config)
	var (
		scheduleQ queue.ScheduleQueue
		binder    binder.Binder
	)
	scheduleQ = kubequeue.NewKubeQueue()
	binder = kubebinder.NewKubeBind()
	for {
		t := scheduleQ.GetTask()
		if t != nil {
			log.Printf("unscheduled pod: %s\n", t.Name)
			if t.Name == "pause" {
				err := binder.Bind(t, "k8s-master01")
				if err != nil {
					log.Fatal(err)
				} else {
					log.Printf("scheduling pod: %s\n", t.Name)
				}
			}
		}
	}
}

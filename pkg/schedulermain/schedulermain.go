package schedulermain

import (
	"github.com/AgentGuo/scheduler/cmd/scheduler-main/config"
	"github.com/AgentGuo/scheduler/task/kubequeue"
	"log"
)

func RunSchedulerMain(config config.Config) {
	log.Println("hello, i am scheduler main! this is the config file")
	log.Printf("%v\n", config)
	q := kubequeue.NewKubeQueue()
	for {
		t := q.GetTask()
		if t != nil {
			log.Printf("schedule pod: %s\n", t.Name)
		}
	}
}

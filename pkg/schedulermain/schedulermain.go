package schedulermain

import (
	"github.com/AgentGuo/scheduler/cmd/scheduler-main/config"
	"log"
)

func RunSchedulerMain(config config.Config) {
	log.Println("hello, i am scheduler main! this is the config file")
	log.Printf("%v\n", config)
}

package app

import (
	"github.com/AgentGuo/scheduler/pkg/schedulermain/task"
	"github.com/AgentGuo/scheduler/pkg/schedulermain/task/resourcequeue"
	"log"
	"net/rpc"
)

func SubmitResourceTask(t task.Task) error {
	client, err := rpc.Dial("tcp", "0.0.0.0:12352")
	if err != nil {
		log.Println("dial error:", err)
		return err
	}
	reply := &resourcequeue.ResourceQueueServiceReply{}
	err = client.Call(resourcequeue.ResourceQueueServiceName+"."+"SubmitTask",
		resourcequeue.ResourceQueueServiceArg{Task: t}, reply)
	if err != nil {
		log.Println("call error:", err)
		return err
	}
	client.Close()
	return nil
}

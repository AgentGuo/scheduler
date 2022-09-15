package resourcequeue

import (
	"github.com/AgentGuo/scheduler/pkg/schedulermain/task"
)

const ResourceQueueServiceName = "ResourceQueueService"

type ResourceQueueService struct {
}

type ResourceQueueServiceArg struct {
	Task task.Task
}

type ResourceQueueServiceReply struct {
}

func (r *ResourceQueueService) SubmitTask(arg ResourceQueueServiceArg, reply *ResourceQueueServiceReply) error {
	return rq.SubmitTask(arg.Task)
}

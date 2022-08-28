package resourcequeue

import (
	"context"
	"sync"

	"github.com/AgentGuo/scheduler/pkg/schedulermain/task"
	"github.com/AgentGuo/scheduler/pkg/schedulermain/task/queue"
	"github.com/AgentGuo/scheduler/util"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Entry

type ResourceQueue struct {
	rw *sync.RWMutex
	q  *queue.TaskQueue
}

func NewResourceQueue(ctx context.Context) *ResourceQueue {
	logger, _ = util.GetCtxLogger(ctx)
	return &ResourceQueue{
		rw: &sync.RWMutex{},
		q:  &queue.TaskQueue{},
	}
}

func (r ResourceQueue) GetTask() *task.Task {
	r.rw.RLock()
	defer r.rw.RUnlock()
	if r.q.Len() == 0 {
		return nil
	} else {
		return (r.q.Pop()).(*task.Task)
	}
}

func (r ResourceQueue) SubmitTask(task task.Task) error {
	r.rw.RLock()
	defer r.rw.RUnlock()

	// TODO: task检查
	r.q.Push(task)
	return nil
}

package resourcequeue

import (
	"context"
	"fmt"
	"net"
	"net/rpc"
	"sync"

	"github.com/AgentGuo/scheduler/pkg/schedulermain/task"
	"github.com/AgentGuo/scheduler/pkg/schedulermain/task/queue"
)

var rq *ResourceQueue

type ResourceQueue struct {
	rw *sync.RWMutex
	q  *queue.TaskQueue
}

func NewResourceQueue(ctx context.Context, port int) (*ResourceQueue, error) {
	rq = &ResourceQueue{
		rw: &sync.RWMutex{},
		q:  &queue.TaskQueue{},
	}
	service := &ResourceQueueService{}
	//service := &HelloService{}
	err := rpc.Register(service)
	if err != nil {
		return nil, err
	}
	listener, err := net.Listen("tcp", "0.0.0.0:"+fmt.Sprintf("%d", port))
	if err != nil {
		return nil, err
	}
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				continue
			}
			go rpc.ServeConn(conn)
		}
	}()
	return rq, nil
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

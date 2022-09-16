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
	index int
	mutex *sync.Mutex
	q     *queue.TaskQueue
}

func AppendResourceQueue(ctx context.Context, list *queue.ScheduleQueueList, port int) error {
	index := len(*list)
	rq = &ResourceQueue{
		index: index,
		mutex: &sync.Mutex{},
		q:     &queue.TaskQueue{},
	}
	service := &ResourceQueueService{}
	err := rpc.Register(service)
	if err != nil {
		return err
	}
	listener, err := net.Listen("tcp", "0.0.0.0:"+fmt.Sprintf("%d", port))
	if err != nil {
		return err
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
	*list = append(*list, rq)
	return nil
}

func (r ResourceQueue) GetTask() *task.Task {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if r.q.Len() == 0 {
		return nil
	} else {
		return (r.q.Pop()).(*task.Task)
	}
}

func (r ResourceQueue) SubmitTask(task task.Task) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// TODO: task检查
	r.q.Push(task)
	queue.TaskChan <- r.index
	return nil
}

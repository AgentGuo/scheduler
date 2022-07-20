package queue

import (
	"github.com/AgentGuo/scheduler/task"
)

// TaskQueue priority queue for task scheduling, it is not thread safe
type TaskQueue []task.Task

func (t TaskQueue) Len() int {
	return len(t)
}

func (t TaskQueue) Less(i, j int) bool {
	return t[i].Priority < t[j].Priority
}

func (t TaskQueue) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func (t *TaskQueue) Push(x interface{}) {
	*t = append(*t, x.(task.Task)) // if x is not Task type, panic will be triggered here
}

func (t *TaskQueue) Pop() interface{} {
	l := len(*t)
	if l <= 0 { // queue is empty
		return nil
	}
	popItem := (*t)[l-1]
	*t = (*t)[:l-1]
	return &popItem
}

// ScheduleQueue general scheduling queue interface
type ScheduleQueue interface {
	GetTask() *task.Task
	SubmitTask(task.Task)
}

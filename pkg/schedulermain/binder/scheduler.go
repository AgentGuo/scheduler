package binder

import "github.com/AgentGuo/scheduler/task"

type Binder interface {
	Bind(t *task.Task, nodeName string) error
}

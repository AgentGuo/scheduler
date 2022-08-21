package schedulermain

import (
	"github.com/AgentGuo/scheduler/pkg/schedulermain/task"
)

type Binder interface {
	Bind(t *task.Task, nodeName string) error
}

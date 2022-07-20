package manage

import (
	"fmt"
	"log"

	"github.com/AgentGuo/scheduler/pkg/resourcemanage/apis"
	"github.com/AgentGuo/scheduler/task"
)

type Manager struct {
	PlatformNotifier *PlatformNotifier
}

func (m *Manager) changeResourceLimitInOthers() error {
	log.Println("Need to implement.")
	return fmt.Errorf("unknown type of task")
}

func (m *Manager) ChangeResourceLimit(args *apis.ResourceModifyArgs, reply *apis.ResourceModifyReply) (err error) {
	switch args.Type {
	case task.KubeResourceTaskType:
		kubeTask := &apis.KubeResourceTask{
			PodName: args.PodName,
			PodUid:  args.PodUid,
			ResourceTask: apis.ResourceTask{
				ContainerName: args.ContainerName,
				ContainerId:   args.ContainerId,
				ResourceValue: apis.ResourceValue{
					CpuLimit:    args.CpuLimit,
					MemoryLimit: args.MemoryLimit,
				},
			},
		}
		err = m.changeResourceLimitInKube(kubeTask)
		if err != nil {
			log.Printf("ChangeResourceLimit failed with task %v", kubeTask)
			reply.Done = false
			return err
		}
		reply.Done = true
		return nil
	default:
		err = m.changeResourceLimitInOthers()
		if err != nil {
			reply.Done = false
			return err
		}
		reply.Done = true
		return nil
	}
}

func NewManager() *Manager {
	return &Manager{
		PlatformNotifier: &PlatformNotifier{},
	}
}

package manage

import (
	"log"

	"github.com/AgentGuo/scheduler/pkg/resourcemanage/apis"
	"github.com/AgentGuo/scheduler/pkg/schedulermain/task"
)

type PlatformManager interface {
	changeResource(t interface{}) error
}

type PlatformNotifier interface {
	Notify(t interface{}) error
}

type Manager struct {
	platformManager  PlatformManager
	platformNotifier PlatformNotifier
}

func (m *Manager) ChangeResourceLimit(args *apis.ResourceModifyArgs, reply *apis.ResourceModifyReply) (err error) {
	switch args.Type {
	case task.KubeResourceTaskType:
		kubeTask := apis.KubeResourceTask{
			PodName:   args.PodName,
			PodUid:    args.PodUid,
			Namespace: args.NameSpace,
			ResourceTask: apis.ResourceTask{
				ContainerName: args.ContainerName,
				ContainerId:   args.ContainerId,
				ResourceValue: apis.ResourceValue{
					CpuLimit:    args.CpuLimit,
					MemoryLimit: args.MemoryLimit,
				},
			},
		}
		err = m.platformManager.changeResource(kubeTask)
		if err != nil {
			log.Printf("ChangeResourceLimit failed with task %v", kubeTask)
			reply.Done = false
			return err
		}
		err = m.platformNotifier.Notify(kubeTask)
		if err != nil {
			// 暂不处理
		}
		reply.Done = true
		return nil
	default:
		reply.Done = false
		return nil
	}
}

func NewManager(pm PlatformManager, pn PlatformNotifier) *Manager {
	return &Manager{
		platformManager:  pm,
		platformNotifier: pn,
	}
}

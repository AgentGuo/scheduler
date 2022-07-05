package manage

import (
	"fmt"
	"log"

	"github.com/AgentGuo/scheduler/pkg/resourcemanage/apis"
	"github.com/AgentGuo/scheduler/util"
)

type Manager struct {
	PlatformNotifier *PlatformNotifier
}

func (m *Manager) changeResourceLimitInKube(target *apis.KubeResourceTask) error {
	var oldValueCpu int64
	var oldValueMem int64
	var err error

	if target.CpuLimit != 0 {
		if oldValueCpu, err = changeLimitInKubeByResource("cpu", apis.CpuLimitInUs, target); err != nil {
			log.Println("change cpu limit failed")
			return err
		}
	}

	if target.MemoryLimit != 0 {
		if oldValueMem, err = changeLimitInKubeByResource("memory", apis.MemoryLimitInBytes, target); err != nil {
			log.Println("change memory limit failed")
			target.CpuLimit = oldValueCpu
			if _, err = changeLimitInKubeByResource("cpu", apis.CpuLimitInUs, target); err != nil {
				log.Println("change back cpu limit failed")
				return err
			}
			return err
		}
	}
	log.Printf("cpu: %d -> %d, mem: %d -> %d\n", oldValueCpu, target.CpuLimit, oldValueMem, target.MemoryLimit)
	// TODO: Notify Platform
	return nil
}

func (m *Manager) changeResourceLimitInOthers() error {
	log.Println("Need to implement.")
	return fmt.Errorf("unknown type of task")
}

func (m *Manager) ChangeResourceLimit(args apis.ResourceModifyArgs, reply *apis.ResourceModifyReply) (err error) {
	switch task := args.ResourceTask.(type) {
	case apis.KubeResourceTask:
		err = m.changeResourceLimitInKube(&task)
	case apis.ResourceTask:
	default:
		err = m.changeResourceLimitInOthers()
	}
	return err
}

func changeLimitInKubeByResource(resource string, changeFile string, target *apis.KubeResourceTask) (int64, error) {
	// check Burstable Pod firstly
	path := target.KubeResourcePathByPodContainerID(resource, apis.KubeBurstableDir, apis.KubeBurstablePodDirPrefix, changeFile)
	var oldValue int64 = 0
	var errw error = nil
	var changeData int64
	if resource == "cpu" {
		changeData = target.CpuLimit
	} else if resource == "memory" {
		changeData = target.MemoryLimit
	} else {
		return 0, fmt.Errorf("wrong resource")
	}
	if ok, err := util.IsDirOrFileExist(path); ok {
		if oldValue, errw = util.WriteIntToFile(path, changeData); errw != nil {
			return oldValue, errw
		} else {
			log.Printf("Modify %s limit success.\n", resource)
		}
	} else {
		if err != nil {
			return oldValue, err
		}
		// then check Besteffort Pod
		path = target.KubeResourcePathByPodContainerID(resource, apis.KubeBesteffortDir, apis.KubeBesteffortPodDirPrefix, changeFile)
		if ok1, err1 := util.IsDirOrFileExist(path); ok1 {
			if oldValue, errw = util.WriteIntToFile(path, changeData); errw != nil {
				return oldValue, errw
			} else {
				log.Printf("Modify %s limit success.\n", resource)
			}
		} else {
			if err1 != nil {
				return oldValue, err
			} else {
				return oldValue, fmt.Errorf("please check podUID and containerID, because file:%s is not exist", path)
			}
		}
	}
	return oldValue, nil
}

func NewManager() *Manager {
	return &Manager{
		PlatformNotifier: &PlatformNotifier{},
	}
}

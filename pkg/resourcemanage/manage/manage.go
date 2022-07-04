package manage

import (
	"fmt"
	"log"

	"github.com/AgentGuo/scheduler/pkg/resourcemanage/apis"
	"github.com/AgentGuo/scheduler/util"
)

type Manager struct {
}

func (m *Manager) changeResourceLimitInKube(target *apis.KubeResourceTask) error {
	cpu := target.CpuLimit
	mem := target.MemoryLimit

	if cpu != 0 {
		return changeLimitInKubeByResource("cpu", cpu, target)
	}

	if mem != 0 {
		return changeLimitInKubeByResource("memory", mem, target)
	}
	log.Println("Nothing changed.")
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

func changeLimitInKubeByResource(resource string, changeData int64, target *apis.KubeResourceTask) error {
	// check Burstable Pod firstly
	path := target.KubeResourcePathByPodContainerID(resource, apis.KubeBurstableDir, apis.KubeBurstablePodDirPrefix, apis.CpuLimitInUs)
	log.Println(path)
	if ok, err := util.IsDirOrFileExist(path); ok {
		if errw := util.WriteIntToFile(path, changeData); errw != nil {
			return errw
		} else {
			log.Printf("Modify %s limit success.\n", resource)
		}
	} else {
		if err != nil {
			return err
		}
		// then check Besteffort Pod
		path = target.KubeResourcePathByPodContainerID(resource, apis.KubeBesteffortDir, apis.KubeBesteffortPodDirPrefix, apis.CpuLimitInUs)
		if ok1, err1 := util.IsDirOrFileExist(path); ok1 {
			if errw := util.WriteIntToFile(path, changeData); errw != nil {
				return errw
			} else {
				log.Printf("Modify %s limit success.\n", resource)
			}
		} else {
			if err1 != nil {
				return err
			} else {
				return fmt.Errorf("please check podUID and containerID, because file:%s is not exist", path)
			}
		}
	}
	return nil
}

func NewManager() *Manager {
	return &Manager{}
}

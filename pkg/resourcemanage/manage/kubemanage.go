package manage

import (
	"fmt"
	"log"

	"github.com/AgentGuo/scheduler/pkg/resourcemanage/apis"
	"github.com/AgentGuo/scheduler/util"
)

type KubeManager struct {
}

func (km KubeManager) changeResource(t interface{}) error {
	target, ok := t.(apis.KubeResourceTask)
	if !ok {
		return fmt.Errorf("target cannot convert to KubeResourceTask")
	}
	var oldValueCpu int64
	var oldValueMem int64
	var err error

	if target.CpuLimit != 0 {
		if oldValueCpu, err = changeLimitInKubeByResource("cpu", apis.CpuLimitInUs, &target); err != nil {
			log.Println("change cpu limit failed")
			return err
		}
	}

	if target.MemoryLimit != 0 {
		if oldValueMem, err = changeLimitInKubeByResource("memory", apis.MemoryLimitInBytes, &target); err != nil {
			log.Println("change memory limit failed")
			target.CpuLimit = oldValueCpu
			if _, err = changeLimitInKubeByResource("cpu", apis.CpuLimitInUs, &target); err != nil {
				log.Println("change back cpu limit failed")
				return err
			}
			return err
		}
	}
	log.Printf("cpu: %d -> %d, mem: %d -> %d\n", oldValueCpu, target.CpuLimit, oldValueMem, target.MemoryLimit)
	return nil
}

func changeLimitInKubeByResource(resource string, changeFile string, target *apis.KubeResourceTask) (int64, error) {
	// check Burstable Pod firstly
	path := target.KubeResourcePathByPodContainerID(resource, apis.KubeBurstableDir, apis.KubeBurstablePodDirPrefix, changeFile)
	var oldValue int64 = 0
	var errw error = nil
	var changeData int64
	if resource == "cpu" {
		changeData = target.CpuLimit * 100 // CpuLimit(m) / 1000m * 100000us(cpu.cfs_period_us)
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

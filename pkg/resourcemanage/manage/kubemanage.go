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

	/*	整体:
		   				   |--fail--> cpu --> |
		cpu --success--> memory --success--> return
		 |                                    |
		 | --fail---------------------------->|

		修改cpu or mem:
		 1.读取旧值
		 2.计算差值
		 3.修改Pod Limit
		 	1)成功.
				修改Container Limit
				a. 成功, 返回
				b. 失败, Pod Limit还原, 返回
			2)失败, return

		只修改了pod level和container level, 高级level均为-1
	*/
	if target.CpuLimit != 0 {
		if oldValueCpu, err = changeLimitInContainer("cpu", apis.CpuLimitInUs, &target); err != nil {
			log.Println("change cpu limit failed")
			return err
		}
	}

	if target.MemoryLimit != 0 {
		if oldValueMem, err = changeLimitInContainer("memory", apis.MemoryLimitInBytes, &target); err != nil {
			log.Println("change memory limit failed")
			if target.CpuLimit != 0 {
				target.CpuLimit = oldValueCpu
				if _, errb := changeLimitInContainer("cpu", apis.CpuLimitInUs, &target); errb != nil {
					log.Println("change back cpu limit failed")
					return errb
				}
			}
			return err
		}
	}
	log.Printf("cpu: %dm -> %dm, mem: %d bytes -> %d bytes\n", oldValueCpu/100, target.CpuLimit, oldValueMem, target.MemoryLimit)
	return nil
}

func changeLimitInContainer(resource string, changeFile string, target *apis.KubeResourceTask) (int64, error) {
	var oldValue int64 = 0
	var errW error = nil
	var errR error = nil
	var changeData int64
	var diff int64
	if resource == "cpu" {
		// the actual admin directory
		resource = "cpu,cpuacct"
		if target.CpuLimit != -1 {
			changeData = target.CpuLimit * 100 // CpuLimit(m) / 1000m * 100000us(cpu.cfs_period_us)
		} else {
			changeData = -1
		}
	} else if resource == "memory" {
		changeData = target.MemoryLimit
	} else {
		return 0, fmt.Errorf("wrong resource")
	}

	switch target.Qos {
	case apis.PodQOSGuaranteed:
		pathC := target.KubeContainerPathByPodContainerID(resource, apis.GuaranteedDir, apis.GuaranteedPodDirPrefix, changeFile)
		pathP := target.KubePodPathByPodID(resource, apis.GuaranteedDir, apis.GuaranteedPodDirPrefix, changeFile)
		if ok, err := util.IsDirOrFileExist(pathC); ok {
			oldValue, diff, errR = checkChange(pathC, changeData)
			if errR != nil {
				return oldValue, errR
			}

			if oldValue == changeData {
				return oldValue, nil
			}

			errW = changeT(diff, oldValue, changeData, pathP, pathC)
			if errW != nil {
				return oldValue, errW
			}
			log.Printf("modify %s limit success.\n", resource)
			break
		} else {
			if err != nil {
				return oldValue, err
			}
			return oldValue, fmt.Errorf("please check podUID and containerID, because file:%s is not exist", pathC)
		}
	case apis.PodQOSBestEffort:
		pathC := target.KubeContainerPathByPodContainerID(resource, apis.KubeBesteffortDir, apis.KubeBesteffortPodDirPrefix, changeFile)
		pathP := target.KubePodPathByPodID(resource, apis.KubeBesteffortDir, apis.KubeBesteffortPodDirPrefix, changeFile)
		if ok, err := util.IsDirOrFileExist(pathC); ok {
			oldValue, diff, errR = checkChange(pathC, changeData)
			if errR != nil {
				return oldValue, errR
			}

			if oldValue == changeData {
				return oldValue, nil
			}

			errW = changeT(diff, oldValue, changeData, pathP, pathC)
			if errW != nil {
				return oldValue, errW
			}
			log.Printf("modify %s limit success.\n", resource)
			break
		} else {
			if err != nil {
				return oldValue, err
			}
			return oldValue, fmt.Errorf("please check podUID and containerID, because file:%s is not exist", pathC)
		}
	case apis.PodQOSBurstable:
		pathC := target.KubeContainerPathByPodContainerID(resource, apis.KubeBurstableDir, apis.KubeBurstablePodDirPrefix, changeFile)
		pathP := target.KubePodPathByPodID(resource, apis.KubeBurstableDir, apis.KubeBurstablePodDirPrefix, changeFile)
		if ok, err := util.IsDirOrFileExist(pathC); ok {
			oldValue, diff, errR = checkChange(pathC, changeData)
			if errR != nil {
				return oldValue, errR
			}

			if oldValue == changeData {
				return oldValue, nil
			}

			errW = changeT(diff, oldValue, changeData, pathP, pathC)
			if errW != nil {
				return oldValue, errW
			}
			log.Printf("modify %s limit success.\n", resource)
			break
		} else {
			if err != nil {
				return oldValue, err
			}
			return oldValue, fmt.Errorf("please check podUID and containerID, because file:%s is not exist", pathC)
		}
	default:
		return -1, fmt.Errorf("qos class error whit %s", target.Qos)
	}
	return oldValue, nil
}

func changeT(diff int64, oldValue int64, changeData int64, pathP string, pathC string) error {
	if diff >= 0 {
		oldP, errW := changeLimitInPod(pathP, diff, false)
		if errW != nil {
			return errW
		}

		if errW = util.WriteIntToFile(pathC, changeData); errW != nil {
			_, errc := changeLimitInPod(pathP, oldP, true)
			return fmt.Errorf("container limit failed [%v] and Pod Limit back [%v]", errW, errc)
		}
	} else {
		if errW := util.WriteIntToFile(pathC, changeData); errW != nil {
			return fmt.Errorf("container limit failed [%v]", errW)
		}

		_, errW := changeLimitInPod(pathP, diff, false)
		if errW != nil {
			if errW = util.WriteIntToFile(pathC, oldValue); errW != nil {
				return fmt.Errorf("container limit back failed [%v]", errW)
			}
			return errW
		}
	}

	return nil
}

func changeLimitInPod(path string, data int64, mode bool) (int64, error) {
	oldValue, errR := util.ReadIntFromFile(path)
	if errR != nil {
		return oldValue, errR
	}

	// 原来是-1, 则不作改变
	if !mode && oldValue == -1 {
		return oldValue, nil
	}

	var newValue int64 = -1
	// mode == true, change back.
	if mode {
		newValue = data
	} else {
		// data = 0, 表示需要改为无限制
		if data != 0 {
			newValue = oldValue + data
		}
	}

	errW := util.WriteIntToFile(path, newValue)
	return oldValue, errW
}

func checkChange(path string, changeData int64) (oldValue int64, diff int64, errR error) {
	// 先读取原值, 计算差值之后, 先修改上一层, 再修改Container层
	oldValue, errR = util.ReadIntFromFile(path)
	if errR != nil {
		return oldValue, diff, errR
	}

	// 改为-1 或 本来就是-1, Pod limit都应该为-1
	if changeData == -1 || oldValue == -1 {
		diff = 0
	} else {
		diff = changeData - oldValue
	}
	return oldValue, diff, nil
}

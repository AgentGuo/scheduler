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
	var oldP int64 = 0
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

	// check Burstable Pod firstly
	path := target.KubeContainerPathByPodContainerID(resource, apis.KubeBurstableDir, apis.KubeBurstablePodDirPrefix, changeFile)
	if ok, err := util.IsDirOrFileExist(path); ok {
		// 先读取原值, 计算差值之后, 先修改Pod层, 再修改Container层
		oldValue, errR = util.ReadIntFromFile(path)
		if errR != nil {
			return oldValue, errR
		}

		if oldValue == changeData {
			return oldValue, nil
		}

		// 改为-1 或 本来就是-1, Pod limit都应该为-1
		if changeData == -1 || oldValue == -1 {
			diff = 0
		} else {
			diff = changeData - oldValue
		}
		oldP, errW = changeLimitInPod(target.KubePodPathByPodID(resource, apis.KubeBurstableDir, apis.KubeBurstablePodDirPrefix, changeFile), diff, false)
		if errW != nil {
			return oldValue, errW
		}

		if errW = util.WriteIntToFile(path, changeData); errW != nil {
			_, errc := changeLimitInPod(target.KubePodPathByPodID(resource, apis.KubeBurstableDir, apis.KubeBurstablePodDirPrefix, changeFile), oldP, true)
			return oldValue, fmt.Errorf("container limit failed [%v] and Pod Limit back [%v]", errW, errc)
		} else {
			log.Printf("modify %s limit success.\n", resource)
		}
	} else {
		if err != nil {
			return oldValue, err
		}
		// then check Besteffort Pod
		path = target.KubeContainerPathByPodContainerID(resource, apis.KubeBesteffortDir, apis.KubeBesteffortPodDirPrefix, changeFile)
		if ok1, err1 := util.IsDirOrFileExist(path); ok1 {
			oldValue, errR = util.ReadIntFromFile(path)
			if errR != nil {
				return oldValue, errR
			}

			if oldValue == changeData {
				return oldValue, nil
			}

			if changeData == -1 || oldValue == -1 {
				diff = 0
			} else {
				diff = changeData - oldValue
			}

			oldP, errW = changeLimitInPod(target.KubePodPathByPodID(resource, apis.KubeBesteffortDir, apis.KubeBesteffortPodDirPrefix, changeFile), diff, false)
			if errW != nil {
				return oldValue, errW
			}

			if errW = util.WriteIntToFile(path, changeData); errW != nil {
				_, errc := changeLimitInPod(target.KubePodPathByPodID(resource, apis.KubeBesteffortDir, apis.KubeBesteffortPodDirPrefix, changeFile), oldP, true)
				return oldValue, fmt.Errorf("container limit failed [%v] and Pod Limit back [%v]", errW, errc)
			} else {
				log.Printf("modify %s limit success.\n", resource)
			}
		} else {
			if err1 != nil {
				return oldValue, err
			} else {
				// finally check Guaranteed Pod
				path = target.KubeContainerPathByPodContainerID(resource, apis.GuaranteedDir, apis.GuaranteedPodDirPrefix, changeFile)
				if ok1, err2 := util.IsDirOrFileExist(path); ok1 {
					oldValue, errR = util.ReadIntFromFile(path)
					if errR != nil {
						return oldValue, errR
					}

					if oldValue == changeData {
						return oldValue, nil
					}

					if changeData == -1 || oldValue == -1 {
						diff = 0
					} else {
						diff = changeData - oldValue
					}

					oldP, errW = changeLimitInPod(target.KubePodPathByPodID(resource, apis.GuaranteedDir, apis.GuaranteedPodDirPrefix, changeFile), diff, false)
					if errW != nil {
						return oldValue, errW
					}

					if errW = util.WriteIntToFile(path, changeData); errW != nil {
						_, errc := changeLimitInPod(target.KubePodPathByPodID(resource, apis.GuaranteedDir, apis.GuaranteedPodDirPrefix, changeFile), oldP, true)
						return oldValue, fmt.Errorf("container limit failed [%v] and Pod Limit back [%v]", errW, errc)
					} else {
						log.Printf("modify %s limit success.\n", resource)
					}
				} else {
					if err2 != nil {
						return oldValue, err2
					} else {
						return oldValue, fmt.Errorf("please check podUID and containerID, because file:%s is not exist", path)
					}
				}
			}
		}
	}
	return oldValue, nil
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

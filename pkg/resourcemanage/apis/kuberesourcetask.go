package apis

import (
	"path/filepath"
	"strings"

	"github.com/AgentGuo/scheduler/util"
)

const (
	CgroupKubeDir              = "kubepods.slice"
	KubeBesteffortDir          = "kubepods-besteffort.slice"
	KubeBurstableDir           = "kubepods-burstable.slice"
	GuaranteedDir              = ""
	KubeBesteffortPodDirPrefix = "kubepods-besteffort-pod"
	KubeBurstablePodDirPrefix  = "kubepods-burstable-pod"
	GuaranteedPodDirPrefix     = "kubepods-pod"
	KubePodDirSuffix           = ".slice"
	KubeDockerDirPrefix        = "docker-"
	KubeDockerDirSuffix        = ".scope"

	// Pod QosClass
	PodQOSGuaranteed = "Guaranteed"
	PodQOSBurstable  = "Burstable"
	PodQOSBestEffort = "BestEffort"
)

type KubeResourceTask struct {
	PodName   string `json:"PodName" yaml:"PodName"`
	PodUid    string `json:"PodUid" yaml:"PodUid"`
	Namespace string `json:"Namespace" yaml:"Namespace"`
	Qos       string `json:"Qos" yaml:"Qos"`
	ResourceTask
}

func (krt KubeResourceTask) KubeContainerPathByPodContainerID(resourceType, kubeQosClassDir, podDirPrefix, resourceLimitFile string) string {
	return filepath.Join(util.CgroupDir(), resourceType, CgroupKubeDir, kubeQosClassDir, util.JoinPath(podDirPrefix, krt.PodUid, KubePodDirSuffix),
		strings.Join([]string{KubeDockerDirPrefix, krt.ContainerId, KubeDockerDirSuffix}, ""), resourceLimitFile)
}

func (krt KubeResourceTask) KubePodPathByPodID(resourceType, kubeQosClassDir, podDirPrefix, resourceLimitFile string) string {
	return filepath.Join(util.CgroupDir(), resourceType, CgroupKubeDir, kubeQosClassDir, util.JoinPath(podDirPrefix, krt.PodUid, KubePodDirSuffix), resourceLimitFile)
}

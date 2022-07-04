package apis

import (
	"path/filepath"
	"strings"

	"github.com/AgentGuo/scheduler/util"
)

const CgroupKubeDir = "kubepods.slice"
const KubeBesteffortDir = "kubepods-besteffort.slice"
const KubeBurstableDir = "kubepods-burstable.slice"
const KubeBesteffortPodDirPrefix = "kubepods-besteffort-pod"
const KubeBurstablePodDirPrefix = "kubepods-burstable-pod"
const KubePodDirSuffix = ".slice"
const KubeDockerDirPrefix = "docker-"
const KubeDockerDirSuffix = ".scope"

type KubeResourceTask struct {
	PodName string
	PodUid  string
	ResourceTask
}

func (krt KubeResourceTask) KubeResourcePathByPodContainerID(resourceType, kubeQosClassDir, podDirPrefix, resourceLimitFile string) string {
	return filepath.Join(util.CgroupDir(), resourceType, CgroupKubeDir, kubeQosClassDir, util.JoinPath(podDirPrefix, krt.PodUid, KubePodDirSuffix),
		strings.Join([]string{KubeDockerDirPrefix, krt.ContainerId, KubeDockerDirSuffix}, ""), resourceLimitFile)
}

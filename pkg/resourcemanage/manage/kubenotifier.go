package manage

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/AgentGuo/scheduler/pkg/resourcemanage/apis"
	"github.com/AgentGuo/scheduler/util"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/retry"
)

type KubeNotifier struct {
}

func (kn KubeNotifier) Notify(t interface{}) error {
	target, ok := t.(apis.KubeResourceTask)
	if !ok {
		return fmt.Errorf("target cannot convert to KubeResourceTask")
	}

	config, err := clientcmd.BuildConfigFromFlags("", filepath.Join(util.HomeDir(), ".kube", "config"))
	if err != nil {
		return err
	}
	// config.APIPath = "api"                       // pods
	// config.GroupVersion = &v1.SchemeGroupVersion // group:"", version: "v1"
	// config.NegotiatedSerializer = scheme.Codecs

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	podClient := clientset.CoreV1().Pods(target.Namespace)
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		pod, err := podClient.Get(context.TODO(), target.PodName, metav1.GetOptions{})
		if err != nil {
			return err
		}

		tmp := target

		for i, c := range pod.Spec.Containers {
			if c.Name != target.ContainerName {
				continue
			}
			if tmp.CpuLimit == -1 {
				tmp.CpuLimit = 0 // k8s里面0表示无限制, 与cgroup不同
			}
			if tmp.MemoryLimit == -1 {
				tmp.MemoryLimit = 0
			}
			// 直接修改, 正确性检查在修改cgroup时完成
			pod.Spec.Containers[i].Resources.Limits[v1.ResourceCPU] = resource.MustParse(strings.Join([]string{strconv.FormatInt(target.CpuLimit, 10), "m"}, ""))             // m
			pod.Spec.Containers[i].Resources.Limits[v1.ResourceMemory] = resource.MustParse(strings.Join([]string{strconv.FormatInt(target.MemoryLimit/1024, 10), "Ki"}, "")) // bytes
			break
		}

		_, updateErr := podClient.Update(context.TODO(), pod, metav1.UpdateOptions{})
		return updateErr
	})
	return retryErr
}

package kubebinder

import (
	"context"
	"fmt"
	"github.com/AgentGuo/scheduler/pkg/schedulermain/task"
	"github.com/AgentGuo/scheduler/pkg/schedulermain/task/kubequeue"
	"github.com/AgentGuo/scheduler/util"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"path/filepath"
)

type kubeBind struct {
	client *kubernetes.Clientset
}

func (k kubeBind) Bind(t *task.Task, nodeName string) error {
	if p, ok := t.Detail.(kubequeue.KubeTaskDetails); ok {
		binding := &v1.Binding{
			ObjectMeta: metav1.ObjectMeta{Namespace: p.Namespace, Name: p.PodName, UID: types.UID(p.UID)},
			Target:     v1.ObjectReference{Kind: "Node", Name: nodeName},
		}
		err := k.client.CoreV1().Pods(binding.Namespace).Bind(context.Background(), binding, metav1.CreateOptions{})
		if err != nil {
			return err
		}
		return nil
	} else {
		return fmt.Errorf("task can not convert to kube task")
	}
}

func NewKubeBind() kubeBind {
	kubeConfig := filepath.Join(util.HomeDir(), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err != nil {
		panic(err.Error())
	}
	client, err := kubernetes.NewForConfig(config)
	return kubeBind{
		client: client,
	}
}

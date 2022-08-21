package kubequeue

import (
	"container/heap"
	"context"
	"fmt"
	"github.com/AgentGuo/scheduler/pkg/schedulermain/task"
	"github.com/AgentGuo/scheduler/pkg/schedulermain/task/queue"
	"github.com/AgentGuo/scheduler/util"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"path/filepath"
	"sync"
	"time"
)

var logger *logrus.Entry

type KubeQueue struct {
	rw *sync.RWMutex
	q  *queue.TaskQueue
}

type KubeTaskDetails struct {
	PodName   string
	Namespace string
	UID       string
}

func NewKubeQueue(ctx context.Context) *KubeQueue {
	// init logger
	logger, _ = util.GetCtxLogger(ctx)
	kq := &KubeQueue{
		rw: &sync.RWMutex{},
		q:  &queue.TaskQueue{},
	}
	heap.Init(kq.q)

	kubeConfig := filepath.Join(util.HomeDir(), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err != nil {
		panic(err.Error())
	}
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	// 初始化informer
	factory := informers.NewSharedInformerFactory(clientSet, time.Minute)
	stopper := make(chan struct{})
	nodeInformer := factory.Core().V1().Pods()
	go factory.Start(stopper)
	informer := nodeInformer.Informer()
	// 添加更新node的handle
	informer.AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			switch t := obj.(type) {
			case *v1.Pod:
				return !assignedPod(t)
			case cache.DeletedFinalStateUnknown:
				if _, ok := t.Obj.(*v1.Pod); ok {
					// The carried object may be stale, so we don't use it to check if
					// it's assigned or not.
					return true
				}
				logger.Debugf("unable to convert object %T to *v1.Pod", obj)
				return false
			default:
				logger.Debugf("unable to convert object %T to *v1.Pod", obj)
				return false
			}
		},
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc:    kq.addPodToSchedulingQueue,
			UpdateFunc: kq.updatePodInSchedulingQueue,
			DeleteFunc: kq.deletePodFromSchedulingQueue,
		},
	})
	go informer.Run(stopper)
	if !cache.WaitForCacheSync(stopper, nodeInformer.Informer().HasSynced) {
		runtime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
	}
	return kq
}

func assignedPod(pod *v1.Pod) bool {
	return len(pod.Spec.NodeName) != 0
}

func (k KubeQueue) GetTask() *task.Task {
	k.rw.RLock()
	defer k.rw.RUnlock()
	if k.q.Len() == 0 {
		return nil
	} else {
		return (k.q.Pop()).(*task.Task)
	}
}

func (k *KubeQueue) addPodToSchedulingQueue(obj interface{}) {
	k.rw.Lock()
	defer k.rw.Unlock()
	pod := obj.(*v1.Pod)
	logger.Debugf("add event for unscheduled pod: %s\n", pod.Name)
	k.q.Push(task.Task{
		Name:   pod.Name + "-" + pod.Namespace,
		Labels: pod.Labels,
		Detail: KubeTaskDetails{
			PodName:   pod.Name,
			Namespace: pod.Namespace,
			UID:       string(pod.UID),
		},
	})
}

func (k *KubeQueue) updatePodInSchedulingQueue(oldObj, newObj interface{}) {
	oldPod, newPod := oldObj.(*v1.Pod), newObj.(*v1.Pod)
	logger.Debugf("update event for pod, old pod: %s, new pod: %s\n", oldPod.Name, newPod.Name)
}

func (k *KubeQueue) deletePodFromSchedulingQueue(obj interface{}) {
	pod := obj.(*v1.Pod)
	logger.Debugf("delete event for pod: %s\n", pod.Name)
}

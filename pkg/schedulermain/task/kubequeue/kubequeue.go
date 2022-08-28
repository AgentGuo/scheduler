package kubequeue

import (
	"container/heap"
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/AgentGuo/scheduler/pkg/resourcemanage/apis"
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
)

var logger *logrus.Entry

type KubeQueue struct {
	rw *sync.RWMutex
	q  *queue.TaskQueue
}

type KubeTaskDetails struct {
	PodName   string `json:"PodName"`
	Namespace string `json:"Namespace"`
	UID       string `json:"UID"`
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

func (k KubeQueue) SubmitTask(task task.Task) error {
	k.rw.RLock()
	defer k.rw.RUnlock()

	k.q.Push(task)
	return nil
}

func (k *KubeQueue) addPodToSchedulingQueue(obj interface{}) {
	k.rw.Lock()
	defer k.rw.Unlock()
	pod := obj.(*v1.Pod)
	logger.Debugf("add event for unscheduled pod: %s\n", pod.Name)
	k.q.Push(task.Task{
		Name:     pod.Name + "-" + pod.Namespace,
		Status:   task.PENDING,
		TaskType: task.NormalTaskType,
		Detail: KubeTaskDetails{
			PodName:   pod.Name,
			Namespace: pod.Namespace,
			UID:       string(pod.UID),
		},
		ResourceDetail: apis.ResourceValue{
			CpuLimit:    getPodCpuLimits(pod),
			MemoryLimit: getPodMemLimits(pod),
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

func getPodCpuLimits(pod *v1.Pod) int64 {
	var cpuLimits int64 = 0
	for _, c := range pod.Spec.Containers {
		cpuLimitQ := c.Resources.Limits[v1.ResourceCPU]
		cpuLimits += int64(cpuLimitQ.AsApproximateFloat64() * 1000) // m
	}

	if cpuLimits == 0 {
		return -1 // no limit, cgroup use -1
	}
	return cpuLimits
}

func getPodMemLimits(pod *v1.Pod) int64 {
	var memLimits int64 = 0
	for _, c := range pod.Spec.Containers {
		memLimitQ := c.Resources.Limits[v1.ResourceMemory]
		memLimits += int64(memLimitQ.AsApproximateFloat64()) // byte
	}

	if memLimits == 0 {
		return -1
	}
	return memLimits
}

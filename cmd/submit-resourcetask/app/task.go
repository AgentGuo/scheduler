package app

import (
	"context"
	"log"
	"net/rpc"
	"path/filepath"
	"strings"

	"github.com/AgentGuo/scheduler/cmd/submit-resourcetask/config"
	"github.com/AgentGuo/scheduler/pkg/resourcemanage/apis"
	"github.com/AgentGuo/scheduler/pkg/schedulermain/task"
	"github.com/AgentGuo/scheduler/pkg/schedulermain/task/resourcequeue"
	"github.com/AgentGuo/scheduler/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func SubmitResourceTask(taskList []task.Task) {
	client, err := rpc.Dial("tcp", "0.0.0.0:12352")
	if err != nil {
		log.Println("dial error:", err)
		return
	}
	reply := &resourcequeue.ResourceQueueServiceReply{}
	for _, t := range taskList {
		err = client.Call(resourcequeue.ResourceQueueServiceName+"."+"SubmitTask",
			resourcequeue.ResourceQueueServiceArg{Task: t}, reply)
		if err != nil {
			log.Println("call error:", err)
		}
	}
	client.Close()
}

func init() {
	kubeConfig := filepath.Join(util.HomeDir(), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err != nil {
		panic(err)
	}
	kubeClient, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
}

var kubeClient *kubernetes.Clientset

func GetTaskList(config *config.ResourceTaskConfig) []task.Task {
	taskList := []task.Task{}
	pod, err := kubeClient.CoreV1().Pods(config.Namespace).Get(context.Background(), config.PodName, metav1.GetOptions{})
	if err != nil {
		panic(err)
	}
	podUid := strings.ReplaceAll(string(pod.UID), "-", "_")
	containerIDMap := map[string]string{}
	for _, v := range pod.Status.ContainerStatuses {
		containerIDMap[v.Name] = v.ContainerID
	}
	log.Printf("%+v\n", containerIDMap)
	for _, v := range config.Containers {
		if containerId, ok := containerIDMap[v.ContainerName]; ok {
			containerId = strings.TrimPrefix(containerId, "docker://")
			detailJson, _ := json.Marshal(apis.KubeResourceTask{
				PodName:   config.PodName,
				PodUid:    podUid,
				Qos:       string(pod.Status.QOSClass),
				Namespace: config.Namespace,
				ResourceTask: apis.ResourceTask{
					ContainerName: v.ContainerName,
					ContainerId:   containerId,
					ResourceValue: apis.ResourceValue{
						CpuLimit:    v.CpuLimit,
						MemoryLimit: v.MemoryLimit,
					},
				},
			})
			taskList = append(taskList, task.Task{
				Name:     config.PodName + "-" + config.Namespace,
				Status:   task.PENDING,
				TaskType: task.KubeResourceTaskType,
				Detail:   string(detailJson),
				NodeName: config.NodeName,
			})
		} else {
			log.Fatalf("can not find container[%+v] in pod[%+v]\n", v.ContainerName, config.PodName)
		}
	}
	return taskList
}

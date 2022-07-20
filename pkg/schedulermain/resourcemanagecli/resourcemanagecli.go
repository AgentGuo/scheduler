package resourcemanagecli

import (
	"fmt"
	"log"
	"net/rpc"
	"strconv"
	"strings"

	"github.com/AgentGuo/scheduler/pkg/resourcemanage/apis"
	"github.com/AgentGuo/scheduler/task"
)

type ResourceClient struct {
	port int
}

func (rc *ResourceClient) Execute(t *task.Task, hostIP string) error {
	args := apis.ResourceModifyArgs{}
	reply := apis.ResourceModifyReply{}

	if t.TaskType == task.KubeResourceTaskType {
		kubeDetail, ok := t.Detail.(apis.KubeResourceTask)
		if !ok {
			return fmt.Errorf("task.Detail can not convert to KubeResourceTask")
		}
		args.Type = task.KubeResourceTaskType
		args.NameSpace = kubeDetail.Namespace
		args.PodName = kubeDetail.PodName
		args.PodUid = kubeDetail.PodUid
		args.ContainerName = kubeDetail.ContainerName
		args.ContainerId = kubeDetail.ContainerId
		args.CpuLimit = kubeDetail.CpuLimit
		args.MemoryLimit = kubeDetail.MemoryLimit
	} else {
		// TODO: other types
		args.Type = task.ResourceTaskType
	}
	err := rc.call(hostIP, apis.ChangeResourceLimit, &args, &reply)
	if err != nil || !reply.Done {
		return fmt.Errorf("%+v excute failed", *t)
	}
	return nil
}

func (rc *ResourceClient) call(host string, rpcName string, args interface{}, reply interface{}) error {
	client, err := rpc.Dial("tcp", strings.Join([]string{host, strconv.Itoa(rc.port)}, ":"))
	if err != nil {
		log.Println("dial error:", err)
		return err
	}
	err = client.Call(apis.ResourceManageRPC+"."+rpcName, args, reply)
	if err != nil {
		log.Println("call error:", err)
		return err
	}
	client.Close()
	return nil
}

func NewResourceClient(port int) *ResourceClient {
	return &ResourceClient{
		port: port,
	}
}
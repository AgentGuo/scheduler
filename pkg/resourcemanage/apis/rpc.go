package apis

import "net/rpc"

const (
	ResourceManageRPC   = "ResourceManageRPC"
	ChangeResourceLimit = "ChangeResourceLimit"
)

type ResourceModifyArgs struct {
	Type          string
	ContainerName string
	ContainerId   string
	CpuLimit      int64
	MemoryLimit   int64

	// k8s
	PodName   string
	PodUid    string
	NameSpace string
}

type ResourceModifyReply struct {
	Done bool
}

type ResourceService interface {
	ChangeResourceLimit(args *ResourceModifyArgs, reply *ResourceModifyReply) error
}

func RegisterService(service ResourceService) error {
	return rpc.RegisterName(ResourceManageRPC, service)
}

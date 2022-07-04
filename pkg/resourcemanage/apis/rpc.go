package apis

import "net/rpc"

const ResourceManageRPC = "ResourceManageRPC"

type ResourceModifyArgs struct {
	ResourceTask interface{}
}

type ResourceModifyReply struct {
	Done bool
}

type ResourceService interface {
	ChangeResourceLimit(args ResourceModifyArgs, reply *ResourceModifyReply) error
}

func RegisterService(service ResourceService) error {
	return rpc.RegisterName(ResourceManageRPC, service)
}

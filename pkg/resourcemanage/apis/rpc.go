package apis

import "net/rpc"

const ResourceManageRPC = "ResourceManageRPC"

type ResourceModifyArgs struct {
	ResourceTask interface{} // TODO: 修改为基本类型
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

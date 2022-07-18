package resourcemanagecli

import (
	"log"
	"net/rpc"

	"github.com/AgentGuo/scheduler/pkg/resourcemanage/apis"
)

func call(rpcName string, args interface{}, reply interface{}) error {
	client, err := rpc.Dial("tcp", "localhost:12351")
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

package resourcemanage

import (
	"fmt"
	"log"
	"net"
	"net/rpc"

	"github.com/AgentGuo/scheduler/cmd/resourcemanage/config"
	"github.com/AgentGuo/scheduler/pkg/resourcemanage/apis"
	"github.com/AgentGuo/scheduler/pkg/resourcemanage/manage"
)

type ResourceManager struct {
	*manage.Manager
}

func NewResourceManager() *ResourceManager {
	return &ResourceManager{
		Manager: manage.NewManager(manage.KubeManager{}, manage.KubeNotifier{}),
	}
}

func (rm *ResourceManager) server(port int) {
	err := apis.RegisterService(rm)
	if err != nil {
		log.Fatal("rpc register error:", err)
		return
	}
	listener, err := net.Listen("tcp", "0.0.0.0:"+fmt.Sprintf("%d", port))
	if err != nil {
		log.Fatal("listen error:", err)
		return
	}
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Fatal("listen accept error:", err)
				continue
			}
			go rpc.ServeConn(conn)
		}
	}()
}

func RunResourceManager(config *config.ResourceManagerConfig) {
	rm := NewResourceManager()
	rm.server(config.Port)

	fmt.Println("[Enter \"q\" to stop]")
	quit := ""
	for quit != "q" {
		fmt.Scanf("%s", &quit)
	}
}

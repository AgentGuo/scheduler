package manage

import (
	"fmt"
	"log"
)

type OthersManager struct {
}

func (m OthersManager) changeResource(t interface{}) error {
	log.Println("Need to implement.")
	return fmt.Errorf("unknown type of task")
}

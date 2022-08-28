package scheduler

import (
	"context"
	"github.com/AgentGuo/scheduler/pkg/schedulermain/task"
	"sync"
)

func (s *Scheduler) filter(ctx context.Context, nodeList []string, t *task.Task) []string {
	m := &sync.Map{}
	for _, node := range nodeList {
		m.Store(node, struct{}{})
	}
	for _, filterP := range s.FilterPlugins {
		// step2 遍历每个节点 运行每个score plugin
		wg := &sync.WaitGroup{}
		for i := range nodeList {
			wg.Add(1)
			go func(nodeName string, group *sync.WaitGroup) {
				defer group.Done()
				ok := filterP.Filter(ctx, nodeName, t)
				if !ok {
					m.Delete(nodeName)
				}
			}(nodeList[i], wg)
		}
		wg.Wait()
		nodeList = []string{}
		m.Range(func(key, value any) bool {
			nodeList = append(nodeList, key.(string))
			return true
		})
	}
	return nodeList
}

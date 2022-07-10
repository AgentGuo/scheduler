package schedule

import (
	"github.com/AgentGuo/scheduler/pkg/metricscli"
	"k8s.io/apimachinery/pkg/util/json"
	"log"
	"sort"
	"sync"
)

type NodeScore struct {
	NodeName string
	Score    float64
}

func (s *Scheduler) score() ([]NodeScore, error) {
	keys, err := s.redisCli.HKeys(metricscli.MetricsInfoKey).Result()
	if err != nil {
		return nil, err
	}
	var m sync.Map
	wg := &sync.WaitGroup{}
	for _, key := range keys {
		wg.Add(1)
		go func(k string, group *sync.WaitGroup) {
			defer group.Done()
			v, err := s.redisCli.HGet(metricscli.MetricsInfoKey, k).Result()
			if err != nil {
				return
			}
			nodeInfo := &metricscli.MetricsInfo{}
			err = json.Unmarshal([]byte(v), nodeInfo)
			if err != nil {
				return
			}
			m.Store(k, nodeScore(nodeInfo))
		}(key, wg)
	}
	wg.Wait()
	priorityList := []NodeScore{}
	m.Range(func(key, value interface{}) bool {
		priorityList = append(priorityList,
			NodeScore{NodeName: key.(string), Score: value.(float64)})
		return true
	})
	sort.SliceStable(priorityList, func(i, j int) bool {
		return priorityList[i].Score > priorityList[j].Score
	})
	log.Printf("%+v\n", priorityList)
	return priorityList, nil
}

func nodeScore(nodeInfo *metricscli.MetricsInfo) float64 {
	return nodeInfo.CpuRemain + nodeInfo.MemFree
}

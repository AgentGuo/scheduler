package scheduler

import (
	"context"
	"github.com/AgentGuo/scheduler/pkg/schedulermain/task"
	"github.com/AgentGuo/scheduler/util"
	"sort"
	"sync"
)

type nodeScore struct {
	NodeName string
	Score    float64
}

func (s *Scheduler) score(ctx context.Context, nodeList []string, t *task.Task) ([]string, error) {
	m := &sync.Map{}
	scoreMap := map[string]float64{}
	for i, scoreP := range s.ScorePlugins {
		// step2 遍历每个节点 运行每个score plugin
		wg := &sync.WaitGroup{}
		for i := range nodeList {
			wg.Add(1)
			go func(nodeName string, group *sync.WaitGroup) {
				defer group.Done()
				nodeScore := scoreP.Score(ctx, nodeName, t)
				m.Store(nodeName, nodeScore)
			}(nodeList[i], wg)
		}
		wg.Wait()
		// step3 分数正则化后按权重聚合
		addScore(m, scoreMap, s.ScoreWeights[i])
	}
	// step4 排序得到调度优先节点列表
	priorityList := []nodeScore{}
	for nodeName, score := range scoreMap {
		priorityList = append(priorityList,
			nodeScore{NodeName: nodeName, Score: score})
	}
	sort.SliceStable(priorityList, func(i, j int) bool {
		return priorityList[i].Score > priorityList[j].Score
	})
	logger, _ := util.GetCtxLogger(ctx)
	logger.Infof("node score rank: %+v", priorityList)
	nodeList = make([]string, len(priorityList))
	for i := range priorityList {
		nodeList[i] = priorityList[i].NodeName
	}
	return nodeList, nil
}

// addScore 聚合每一次插件打分的结果
func addScore(m *sync.Map, scoreMap map[string]float64, weight float64) {
	var sum float64 = 0
	m.Range(func(key, value any) bool {
		sum += value.(float64)
		return true
	})
	// 正则化 + 加权
	m.Range(func(key, value any) bool {
		scoreMap[key.(string)] += (value.(float64) / sum) * weight
		return true
	})
}

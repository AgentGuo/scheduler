package scheduler

import (
	"context"
	"github.com/AgentGuo/scheduler/pkg/metricscli"
	"github.com/AgentGuo/scheduler/pkg/schedulermain/task"
	"github.com/AgentGuo/scheduler/util"
	"sort"
	"sync"
)

type NodeScore struct {
	NodeName string
	Score    float64
}

func (s *Scheduler) score(ctx context.Context, t *task.Task) ([]NodeScore, error) {
	// step1 获取所有节点
	keys, err := s.RedisCli.HKeys(metricscli.MetricsInfoKey).Result()
	if err != nil {
		return nil, err
	}
	m := &sync.Map{}
	scoreMap := map[string]float64{}
	for i, scoreP := range s.ScorePlugin {
		// step2 遍历每个节点 运行每个score plugin
		wg := &sync.WaitGroup{}
		for i := range keys {
			wg.Add(1)
			go func(nodeName string, group *sync.WaitGroup) {
				defer group.Done()
				nodeScore := scoreP.Score(ctx, nodeName, t)
				m.Store(nodeName, nodeScore)
			}(keys[i], wg)
		}
		wg.Wait()
		// step3 分数正则化后按权重聚合
		addScore(m, scoreMap, s.ScoreWeight[i])
	}
	// step4 排序得到调度优先节点列表
	priorityList := []NodeScore{}
	for nodeName, score := range scoreMap {
		priorityList = append(priorityList,
			NodeScore{NodeName: nodeName, Score: score})
	}
	sort.SliceStable(priorityList, func(i, j int) bool {
		return priorityList[i].Score > priorityList[j].Score
	})
	logger, _ := util.GetCtxLogger(ctx)
	logger.Infof("node score rank: %+v", priorityList)
	return priorityList, nil
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

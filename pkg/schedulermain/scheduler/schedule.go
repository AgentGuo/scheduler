package scheduler

import (
	"context"
	"fmt"
	"github.com/AgentGuo/scheduler/cmd/scheduler-main/config"
	"github.com/AgentGuo/scheduler/pkg/metricscli"
	"github.com/AgentGuo/scheduler/pkg/schedulermain/scheduler/plugin"
	"github.com/AgentGuo/scheduler/pkg/schedulermain/task"
	"github.com/go-redis/redis"
	"k8s.io/apimachinery/pkg/util/json"
	"time"
)

type Scheduler struct {
	RedisCli      *redis.Client
	ScorePlugins  []plugin.ScorePlugin
	ScoreWeights  []float64
	FilterPlugins []plugin.FilterPlugin
}

func NewScheduler(cfg *config.SchedulerMainConfig) (*Scheduler, error) {
	redisCli := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
	})
	_, err := redisCli.Ping().Result()
	if err != nil {
		return nil, err
	}
	pluginRegMap := GetRegistryMap()
	scorePlugins, scoreWeights := InitScorePlugin(redisCli, cfg, pluginRegMap)
	filterPlugins := InitFilterPlugin(redisCli, cfg, pluginRegMap)
	return &Scheduler{
		RedisCli:      redisCli,
		ScorePlugins:  scorePlugins,
		ScoreWeights:  scoreWeights,
		FilterPlugins: filterPlugins,
	}, nil
}

func (s *Scheduler) Schedule(ctx context.Context, t *task.Task) (nodeName string, err error) {
	nodeList := ListLiveNode(s.RedisCli)
	nodeList = s.filter(ctx, nodeList, t)
	nodeList, err = s.score(ctx, nodeList, t)
	if err != nil {
		return "", err
	}
	if len(nodeList) == 0 {
		return "", fmt.Errorf("no node for scheduling")
	}
	return nodeList[0], nil
}

func ListLiveNode(client *redis.Client) []string {
	res := []string{}
	keys, err := client.HKeys(metricscli.MetricsInfoKey).Result()
	if err != nil {
		return res
	}
	for _, key := range keys {
		v, err := client.HGet(metricscli.MetricsInfoKey, key).Result()
		if err != nil {
			continue
		}
		nodeInfo := &metricscli.MetricsInfo{}
		err = json.Unmarshal([]byte(v), nodeInfo)
		if err != nil {
			continue
		}
		// 说明主机在线
		if time.Now().Add(-time.Second*5).Unix() < nodeInfo.TimeStamp {
			res = append(res, key)
		}
	}
	return res
}

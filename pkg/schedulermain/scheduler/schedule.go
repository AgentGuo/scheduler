package scheduler

import (
	"context"
	"fmt"
	"github.com/AgentGuo/scheduler/cmd/scheduler-main/config"
	"github.com/AgentGuo/scheduler/pkg/schedulermain/scheduler/plugin"
	"github.com/AgentGuo/scheduler/pkg/schedulermain/task"
	"github.com/go-redis/redis"
)

type Scheduler struct {
	RedisCli    *redis.Client
	ScorePlugin []plugin.ScorePlugin
	ScoreWeight []float64
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
	scorePlugin, scoreWeight := InitScorePlugin(redisCli, cfg, pluginRegMap)
	return &Scheduler{
		RedisCli:    redisCli,
		ScorePlugin: scorePlugin,
		ScoreWeight: scoreWeight,
	}, nil
}

func (s *Scheduler) Schedule(ctx context.Context, t *task.Task) (nodeName string, err error) {
	priorityList, err := s.score(ctx, t)
	if err != nil {
		return "", err
	}
	if len(priorityList) == 0 {
		return "", fmt.Errorf("no node for scheduling")
	}
	return priorityList[0].NodeName, nil
}

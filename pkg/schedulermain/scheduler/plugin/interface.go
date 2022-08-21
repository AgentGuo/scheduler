package plugin

import (
	"context"
	"github.com/AgentGuo/scheduler/pkg/schedulermain/task"
	"github.com/go-redis/redis"
)

const PluginLogKey = "plugin"

type PluginFactory = func(client *redis.Client) Plugin

type Plugin interface {
	Name() string
}

type ScorePlugin interface {
	Plugin
	Score(ctx context.Context, nodeName string, task *task.Task) float64
}

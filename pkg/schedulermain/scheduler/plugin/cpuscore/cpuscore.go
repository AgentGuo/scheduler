package cpuscore

import (
	"context"
	"github.com/AgentGuo/scheduler/pkg/metricscli"
	"github.com/AgentGuo/scheduler/pkg/schedulermain/scheduler/plugin"
	"github.com/AgentGuo/scheduler/pkg/schedulermain/task"
	"github.com/AgentGuo/scheduler/util"
	"github.com/go-redis/redis"
	"k8s.io/apimachinery/pkg/util/json"
)

const PluginName = "cpuScore"

type CpuScore struct {
	redisCli *redis.Client
}

func (b CpuScore) Score(ctx context.Context, nodeName string, task *task.Task) float64 {
	v, err := b.redisCli.HGet(metricscli.MetricsInfoKey, nodeName).Result()
	if err != nil {
		return 0
	}
	nodeInfo := &metricscli.MetricsInfo{}
	err = json.Unmarshal([]byte(v), nodeInfo)
	logger, _ := util.GetCtxLogger(ctx)
	logger.WithField(plugin.PluginLogKey, PluginName).Debugf(
		"plugin [%+v]: score-%d", PluginName, nodeInfo.CpuRemain)
	return float64(nodeInfo.CpuRemain)
}

func (b CpuScore) Name() string {
	return PluginName
}

func New(client *redis.Client) plugin.Plugin {
	return &CpuScore{
		redisCli: client,
	}
}

package nodeselector

import (
	"context"
	"github.com/AgentGuo/scheduler/pkg/metricscli"
	"github.com/AgentGuo/scheduler/pkg/schedulermain/scheduler/plugin"
	"github.com/AgentGuo/scheduler/pkg/schedulermain/task"
	"github.com/AgentGuo/scheduler/util"
	"github.com/go-redis/redis"
	"k8s.io/apimachinery/pkg/util/json"
)

const PluginName = "nodeSelector"

type NodeSelector struct {
	redisCli *redis.Client
}

func (n NodeSelector) Name() string {
	return PluginName
}

func (n NodeSelector) Filter(ctx context.Context, nodeName string, t *task.Task) bool {
	res := true
	v, err := n.redisCli.HGet(metricscli.MetricsInfoKey, nodeName).Result()
	if err != nil {
		return false
	}
	nodeInfo := &metricscli.MetricsInfo{}
	err = json.Unmarshal([]byte(v), nodeInfo)
	logger, _ := util.GetCtxLogger(ctx)
	for label, value := range t.Labels {
		if v, ok := nodeInfo.Labels[label]; ok {
			if v != value {
				res = false
			}
		} else {
			res = false
		}
	}
	logger.WithField(plugin.PluginLogKey, PluginName).Debugf(
		"plugin [%+v]: filter result-%+v", PluginName, res)
	return res
}

func New(client *redis.Client) plugin.Plugin {
	return &NodeSelector{
		redisCli: client,
	}
}

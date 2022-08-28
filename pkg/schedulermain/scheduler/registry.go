package scheduler

import (
	"github.com/AgentGuo/scheduler/cmd/scheduler-main/config"
	"github.com/AgentGuo/scheduler/pkg/schedulermain/scheduler/plugin"
	"github.com/AgentGuo/scheduler/pkg/schedulermain/scheduler/plugin/cpuscore"
	"github.com/AgentGuo/scheduler/pkg/schedulermain/scheduler/plugin/memscore"
	"github.com/AgentGuo/scheduler/pkg/schedulermain/scheduler/plugin/nodeselector"
	"github.com/go-redis/redis"
)

func InitScorePlugin(client *redis.Client, cfg *config.SchedulerMainConfig,
	registryPlugin map[string]plugin.PluginFactory) ([]plugin.ScorePlugin, []float64) {
	scorePluginList := []plugin.ScorePlugin{}
	scoreWeight := []float64{}
	for _, v := range cfg.Plugin.Score {
		if regFunc, ok := registryPlugin[v.Name]; ok {
			p := regFunc(client)
			if scoreP, ok := p.(plugin.ScorePlugin); ok {
				scorePluginList = append(scorePluginList, scoreP)
				scoreWeight = append(scoreWeight, v.Weight)
			}
		}
	}
	return scorePluginList, scoreWeight
}

func InitFilterPlugin(client *redis.Client, cfg *config.SchedulerMainConfig,
	registryPlugin map[string]plugin.PluginFactory) []plugin.FilterPlugin {
	filterPluginList := []plugin.FilterPlugin{}
	for _, v := range cfg.Plugin.Filter {
		if regFunc, ok := registryPlugin[v.Name]; ok {
			p := regFunc(client)
			if filterP, ok := p.(plugin.FilterPlugin); ok {
				filterPluginList = append(filterPluginList, filterP)
			}
		}
	}
	return filterPluginList
}

func GetRegistryMap() map[string]plugin.PluginFactory {
	return map[string]plugin.PluginFactory{
		cpuscore.PluginName:     cpuscore.New,
		memscore.PluginName:     memscore.New,
		nodeselector.PluginName: nodeselector.New,
	}
}

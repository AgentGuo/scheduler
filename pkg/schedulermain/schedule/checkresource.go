package schedule

import (
	"fmt"

	"github.com/AgentGuo/scheduler/pkg/metricscli"
)

type CheckReourceInfo struct {
	OldCpuLimit int64
	NewCpuLimit int64
	OldMemLimit int64
	NewMemLimit int64
}

func (s *Scheduler) checkReource(nodeInfo *metricscli.NodeInfo, metricsInfo *metricscli.MetricsInfo, checkReourceInfo *CheckReourceInfo) error {
	if !checkCpu(nodeInfo, metricsInfo, checkReourceInfo) {
		return fmt.Errorf("check Cpu failed")
	}
	if !checkMem(nodeInfo, metricsInfo, checkReourceInfo) {
		return fmt.Errorf("check Memory failed")
	}
	return nil
}

// 暂无实现
func checkCpu(nodeInfo *metricscli.NodeInfo, metricsInfo *metricscli.MetricsInfo, checkReourceInfo *CheckReourceInfo) bool {

	return true
}

func checkMem(nodeInfo *metricscli.NodeInfo, metricsInfo *metricscli.MetricsInfo, checkReourceInfo *CheckReourceInfo) bool {

	return true
}

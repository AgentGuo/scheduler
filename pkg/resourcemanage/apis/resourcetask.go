package apis

const (
	MemoryLimitInBytes = "memory.limit_in_bytes" // -1: no limit
	MemswLimitInByte   = "memory.memsw.limit_in_bytes"

	// cpu.cfs_quota_us / cpu.cfs_period_us = cores(m) can be uesd
	CpuLimitInUs      = "cpu.cfs_quota_us"  // -1: no limit
	CpuPeriodInUs     = "cpu.cfs_period_us" // default :100000(100ms)
	Cpu_cfs_period_us = 100000
)

type ResourceValue struct { // bytes
	CpuLimit    int64 `json:"CpuLimit" yaml:"CpuLimit"`       // 单位为m
	MemoryLimit int64 `json:"MemoryLimit" yaml:"MemoryLimit"` // 单位为byte
}

type ResourceTask struct {
	ContainerName string `json:"ContainerName" yaml:"ContainerName"`
	ContainerId   string `json:"ContainerId" yaml:"ContainerId"`
	ResourceValue
}

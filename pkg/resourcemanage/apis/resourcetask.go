package apis

const MemoryLimitInBytes = "memory.limit_in_bytes"
const CpuLimitInUs = "cpu.cfs_quota_us"

type ResourceValue struct { // bytes
	CpuLimit    int64
	MemoryLimit int64
}

type ResourceTask struct {
	ContainerName string
	ContainerId   string
	ResourceValue
}

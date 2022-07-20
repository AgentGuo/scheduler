package util_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/AgentGuo/scheduler/pkg/resourcemanage/apis"
	"github.com/AgentGuo/scheduler/task"
	"github.com/AgentGuo/scheduler/util"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

func TestJoinPath(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
		id     string
		suffix string
		want   string
	}{
		{"test#1", "besteffort-pod", "abc", ".slice", "besteffort-podabc.slice"},
		{"test#2", "burstable-pod", "abc", ".slice", "burstable-podabc.slice"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := util.JoinPath(tt.prefix, tt.id, tt.suffix); got != tt.want {
				t.Errorf("JoinPath(%s, %s, %s) is %s, want %s", tt.prefix, tt.id, tt.suffix, got, tt.want)
			}
		})
	}
}

func TestIsDirOrFileExist(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{"test#1", "/root/test", true},
		{"test#2", "/root/test_non", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := util.IsDirOrFileExist(tt.path); got != tt.want {
				t.Errorf("IsDirOrFileExist(%s) is %t, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestWriteIntToFile(t *testing.T) {
	tests := []struct {
		name string
		path string
		data int64
		old  int64
		want error
	}{
		{"test#1", "/root/test", 20000, 1, nil},
		{"test#2", "/root/test_non", 1, 0, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if ok, err1 := util.IsDirOrFileExist(tt.path); ok {
				if old, err := util.WriteIntToFile(tt.path, tt.data); err != tt.want {
					t.Errorf("WriteIntToFile(%s, %d) return %d, %s, want %v", tt.path, tt.data, old, err, tt.want)
				}
			} else {
				t.Errorf("error %v", err1)
			}

		})
	}
}

func TestGetLocalIP(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		ip, err := util.GetLocalIP()
		if err == nil {
			t.Errorf(ip)
		}
	})
}

func TestGetNodeInfo(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		coreNums, err := cpu.Counts(true)
		if err != nil {
			t.Errorf("get coreNums failed")
		}
		coreNums *= 1000 // 单位为mCPU
		memStat, err := mem.VirtualMemory()
		if err != nil {
			t.Errorf("get memStat failed")
		}
		memTotal := memStat.Total
		memFree := memStat.Free
		t.Errorf("coreNums: %d, memStat.Total:%d, memTotal: %d, memFree: %d", coreNums, memStat.Total, memTotal, memFree)
	})
}

func TestJson(t *testing.T) {
	taski := task.Task{
		Name:       "pod-namespace",
		Status:     task.RUNNING,
		Priority:   1,
		UpdateTime: time.Now().Unix(),
		TaskType:   task.KubeResourceTaskType,
		NodeName:   "1",
		Detail: apis.KubeResourceTask{
			PodName:   "pod",
			PodUid:    "sadringf",
			Namespace: "namespace",
			ResourceTask: apis.ResourceTask{
				ContainerName: "container",
				ContainerId:   "conID",
				ResourceValue: apis.ResourceValue{
					CpuLimit:    1,
					MemoryLimit: 100,
				},
			},
		},
		ResourceDetail: apis.ResourceValue{
			CpuLimit:    5,
			MemoryLimit: 100,
		},
	}
	t.Run("test", func(t *testing.T) {
		v, err := json.Marshal(taski)
		if err != nil {
			t.Error(err)
		}
		t.Errorf("%v\n", v)
		taskInfo := &task.Task{}
		err = json.Unmarshal([]byte(v), taskInfo) // 解析出来, 接口里面的自定义结构体会转为map
		if err != nil {
			t.Errorf("unmarshal TaskInfo failed\n")
		}
		t.Errorf("%+v\n", *taskInfo)
		// t.Errorf("%+v\n", taskInfo.Detail.(apis.KubeResourceTask))
		// t.Errorf("%+v\n", taskInfo.ResourceDetail.(apis.ResourceValue))
		tm := taskInfo.Detail.(map[string]interface{})
		kubeDetail := apis.KubeResourceTask{
			PodName:   tm["PodName"].(string),
			PodUid:    tm["PodUid"].(string),
			Namespace: tm["PodUid"].(string),
			ResourceTask: apis.ResourceTask{
				ContainerName: tm["ContainerName"].(string),
				ContainerId:   tm["ContainerId"].(string),
				ResourceValue: apis.ResourceValue{
					CpuLimit:    int64(tm["CpuLimit"].(float64)),
					MemoryLimit: int64(tm["MemoryLimit"].(float64)),
				},
			}}
		t.Errorf("%#v\n", kubeDetail)
	})
}

package config

import (
	"testing"
)

func TestReadConfig(t *testing.T) {
	tests := []struct {
		name           string
		configFilepath string
		want           ResourceTaskConfig
	}{
		{"test#1", "./test.yaml", ResourceTaskConfig{
			PodName:       "test",
			PodUid:        "testUid",
			Namespace:     "testspace",
			ContainerName: "testC",
			ContainerId:   "testCUid",
			CpuLimit:      -2,
			MemoryLimit:   -2,
		},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := ReadConfig(tt.configFilepath); err != nil || *got != tt.want {
				t.Errorf("ReadConfig(%s) = %v , want %v, error is %v", tt.configFilepath, *got, tt.want, err)
			}
		})
	}
}

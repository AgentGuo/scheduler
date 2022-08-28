// Package config is config definition for binder-main
package config

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// SchedulerMainConfig binder-main config
type ResourceTaskConfig struct {
	PodName       string `yaml:"PodName"`
	PodUid        string `yaml:"PodUid"`
	Namespace     string `yaml:"Namespace"`
	ContainerName string `yaml:"ContainerName"`
	ContainerId   string `yaml:"ContainerId"`
	CpuLimit      int64  `yaml:"CpuLimit"`
	MemoryLimit   int64  `yaml:"MemoryLimit"`
}

func ReadConfig(configFilePath string) (*ResourceTaskConfig, error) {
	config := &ResourceTaskConfig{ // default config value
		PodName:       "",
		PodUid:        "",
		Namespace:     "",
		ContainerName: "",
		ContainerId:   "",
		CpuLimit:      0,
		MemoryLimit:   0,
	}
	// Must config set
	if len(configFilePath) == 0 {
		return config, fmt.Errorf("no config")
	}
	yamlFile, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return config, err
	}
	// yaml unmarshal
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return config, err
	}
	if config.PodName == "" || config.PodUid == "" || config.Namespace == "" || config.ContainerName == "" || config.ContainerId == "" {
		return config, fmt.Errorf("wrong config")
	}
	return config, nil
}

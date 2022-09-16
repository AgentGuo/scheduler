// Package config is config definition for binder-main
package config

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// SchedulerMainConfig binder-main config
type ResourceTaskConfig struct {
	PodName    string      `yaml:"PodName"`
	Namespace  string      `yaml:"Namespace"`
	NodeName   string      `yaml:"NodeName"`
	Containers []Container `yaml:"Containers"`
}

type Container struct {
	ContainerName string `yaml:"ContainerName"`
	CpuLimit      int64  `yaml:"CpuLimit"`
	MemoryLimit   int64  `yaml:"MemoryLimit"`
}

func ReadConfig(configFilePath string) (*ResourceTaskConfig, error) {
	config := &ResourceTaskConfig{}
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
	if config.PodName == "" || config.Namespace == "" || len(config.Containers) == 0 {
		return config, fmt.Errorf("wrong config")
	}
	return config, nil
}

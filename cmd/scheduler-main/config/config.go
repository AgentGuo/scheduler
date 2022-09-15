// Package config is config definition for binder-main
package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// SchedulerMainConfig binder-main config
type SchedulerMainConfig struct {
	Port                    int          `yaml:"port"` // server port
	RedisAddr               string       `yaml:"redis_addr"`
	RedisPassword           string       `yaml:"redis_password"`
	ResourceManagerPort     int          `yaml:"resource_manager_port"`
	ResourceQueueServerPort int          `yaml:"resource_queue_server_port"`
	LogLevel                string       `yaml:"log_level"`
	Plugin                  PluginConfig `yaml:"plugin"`
}

type PluginConfig struct {
	Filter []*FilterPlugin `yaml:"filter"`
	Score  []*ScorePlugin  `yaml:"score"`
}

type ScorePlugin struct {
	Name   string  `yaml:"name"`
	Weight float64 `yaml:"weight"`
}

type FilterPlugin struct {
	Name string `yaml:"name"`
}

func ReadConfig(configFilePath string) (*SchedulerMainConfig, error) {
	config := &SchedulerMainConfig{ // default config value
		Port:                12345,
		ResourceManagerPort: 12350,
	}
	// No config set, use default
	if len(configFilePath) == 0 {
		return config, nil
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
	return config, nil
}

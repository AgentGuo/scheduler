// Package config is config definition for binder-main
package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

// SchedulerMainConfig binder-main config
type SchedulerMainConfig struct {
	Port          int          `yaml:"port"` // server port
	RedisAddr     string       `yaml:"redis_addr"`
	RedisPassword string       `yaml:"redis_password"`
	LogLevel      string       `yaml:"log_level"`
	Plugin        PluginConfig `yaml:"plugin"`
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
		Port: 12345,
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

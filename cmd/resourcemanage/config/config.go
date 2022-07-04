// Package config is config definition for binder-main
package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// ResourceManagerConfig ResourceManager config
type ResourceManagerConfig struct {
	Port int `yaml:"port"` // server port
}

func ReadConfig(configFilePath string) (*ResourceManagerConfig, error) {
	config := &ResourceManagerConfig{ // default config value
		Port: 12350,
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

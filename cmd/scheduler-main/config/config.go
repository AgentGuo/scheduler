// Package config is config definition for binder-main
package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

// Config binder-main config
type Config struct {
	Port int `yaml:"port"` // server port
}

func ReadConfig(configFilePath string) (Config, error) {
	config := Config{ // default config value
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

package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type MetricsCliConfig struct {
	RedisAddr     string `yaml:"redis_addr"`
	RedisPassword string `yaml:"redis_password"`
}

func ReadConfig(configFilePath string) (*MetricsCliConfig, error) {
	config := &MetricsCliConfig{ // default config value
		RedisAddr:     "127.0.0.1:6379",
		RedisPassword: "foobared",
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

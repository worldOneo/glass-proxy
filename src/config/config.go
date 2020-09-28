package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

// Config the configuration for the ProxyService
type Config struct {
	Addr              string       `json:"addr"`
	Hosts             []HostConfig `json:"hosts"`
	LoggConnections   bool         `json:"logConnections"`
	HealthCheckTime   int          `json:"healthCheckSeconds"`
	SaveConfigOnClose bool         `json:"saveConfigOnClose"`
}

// HostConfig a config for a specific single host
type HostConfig struct {
	Name string `json:"name"`
	Addr string `json:"addr"`
}

// Load loads a config from the
func Load(path string) (*Config, error) {
	configFile, err := os.Open(path)
	defer configFile.Close()
	if err != nil {
		return nil, err
	}
	var config Config
	jsonParser := json.NewDecoder(configFile)
	parseErr := jsonParser.Decode(&config)

	if parseErr != nil {
		return nil, parseErr
	}
	return &config, nil
}

// Create creates a config file
func Create(path string, config *Config) error {
	data, jsonErr := json.MarshalIndent(config, "", "    ")
	if jsonErr != nil {
		return jsonErr
	}
	writingErr := ioutil.WriteFile(path, data, 0644)

	if writingErr != nil {
		return writingErr
	}
	return nil
}

// Default returns a default config
func Default() *Config {
	return &Config{
		Addr: "0.0.0.0:25565",
		Hosts: []HostConfig{
			HostConfig{
				Name: "Server-1",
				Addr: "localhost:25580",
			},
		},
		LoggConnections:   true,
		HealthCheckTime:   5,
		SaveConfigOnClose: false,
	}
}

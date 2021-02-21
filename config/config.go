package config

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
)

// Config the configuration for the ProxyService
type Config struct {
	Protocol          string       `json:"protocol"`
	Addr              string       `json:"addr"`
	Interfaces        []string     `json:"interfaces"`
	Hosts             []HostConfig `json:"hosts"`
	LogConfig         LogConfig    `json:"LogConfiguration"`
	HealthCheckTime   int          `json:"healthCheckSeconds"`
	SaveConfigOnClose bool         `json:"saveConfigOnClose"`
}

// HostConfig a config for a specific single host
type HostConfig struct {
	Name string `json:"name"`
	Addr string `json:"addr"`
}

// LogConfig defines what should be logged and what not
type LogConfig struct {
	LogConnections bool `json:"logConnections"`
	LogDisconnect  bool `json:"logDisconnect"`
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
	config.fillFlags()
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
	conf := &Config{
		Protocol: "tcp",
		Addr:     "0.0.0.0:25565",
		Hosts: []HostConfig{
			{
				Name: "Server-1",
				Addr: "localhost:25580",
			},
		},
		LogConfig: LogConfig{
			LogConnections: true,
			LogDisconnect:  false,
		},
		HealthCheckTime:   5,
		SaveConfigOnClose: false,
		Interfaces:        []string{},
	}
	conf.fillFlags()
	return conf
}

func (c *Config) fillFlags() {
	flag.BoolVar(&c.LogConfig.LogConnections, "logc", c.LogConfig.LogConnections, "Log connections which where successfully bridged.")
	flag.BoolVar(&c.LogConfig.LogDisconnect, "logd", c.LogConfig.LogDisconnect, "Log connections which where closed.")
	flag.BoolVar(&c.SaveConfigOnClose, "save", c.SaveConfigOnClose, "Save the config when the server is stopped.")
	flag.StringVar(&c.Addr, "addr", c.Addr, "The addr to start the server on.")
	flag.IntVar(&c.HealthCheckTime, "health", c.HealthCheckTime, "The time (in seconds) between health checks.")
	flag.Parse()
}

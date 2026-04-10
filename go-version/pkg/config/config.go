package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type TicketConfig struct {
	Username  string `yaml:"username"`
	Password  string `yaml:"password"`
	Realm     string `yaml:"realm"`
	NewFormat string `yaml:"new-format"`
}

type NodeConfig struct {
	Identify string       `yaml:"identify"`
	Host     string       `yaml:"host"`
	Ticket   TicketConfig `yaml:"ticket"`
}

type Config struct {
	Monitoring []NodeConfig `yaml:"monitoring"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

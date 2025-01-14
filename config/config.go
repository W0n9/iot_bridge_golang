package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Sensor struct {
	IP       string `yaml:"ip"`
	Campus   string `yaml:"campus"`
	Building string `yaml:"building"`
	Room     string `yaml:"room"`
}

type Config struct {
	Sensors []Sensor `yaml:"sensors"`
}

func LoadConfig(filename string) (*Config, error) {
	bytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(bytes, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

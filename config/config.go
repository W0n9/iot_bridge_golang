package config

import (
	"fmt"
	"net"
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

// validateIP 检查IP地址格式是否正确
func validateIP(ip string) error {
	if parsedIP := net.ParseIP(ip); parsedIP == nil {
		return fmt.Errorf("invalid IP address: %s", ip)
	}
	return nil
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

	// 验证所有传感器的IP地址
	for _, sensor := range config.Sensors {
		if err := validateIP(sensor.IP); err != nil {
			return nil, fmt.Errorf("sensor validation error: %v", err)
		}
	}

	return &config, nil
}

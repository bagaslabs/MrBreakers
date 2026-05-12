package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Mode        string `json:"mode"`
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Connections int    `json:"connections"`
	IntervalMs  int    `json:"interval_ms"`
	Payload     string `json:"payload"`
}

func LoadConfig(path string) (*Config, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(file, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

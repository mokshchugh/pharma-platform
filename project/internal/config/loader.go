package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func Load(path string) (*Config, error) {
	cfg := &Config{}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open config %s: %w", path, err)
	}
	defer file.Close()

	if err := yaml.NewDecoder(file).Decode(cfg); err != nil {
		return nil, fmt.Errorf("decode config %s: %w", path, err)
	}

	return cfg, nil
}

package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Load reads the complete application configuration from a configuration directory.
func Load(configDir string) (*Config, error) {
	cfg := &Config{}

	if err := loadYAML(filepath.Join(configDir, "plant.yaml"), &cfg.Plant); err != nil {
		return nil, fmt.Errorf("load plant configuration: %w", err)
	}

	if err := loadYAML(filepath.Join(configDir, "collector.yaml"), &cfg.Collector); err != nil {
		return nil, fmt.Errorf("load collector configuration: %w", err)
	}

	if err := loadYAML(filepath.Join(configDir, "api.yaml"), &cfg.API); err != nil {
		return nil, fmt.Errorf("load api configuration: %w", err)
	}

	if err := loadYAML(filepath.Join(configDir, "aggregation.yaml"), &cfg.Aggregator); err != nil {
		return nil, fmt.Errorf("load aggregation configuration: %w", err)
	}

	if err := loadYAML(filepath.Join(configDir, "plcs.yaml"), &cfg.PLCs); err != nil {
		return nil, fmt.Errorf("load plcs configuration: %w", err)
	}

	if err := loadYAML(filepath.Join(configDir, "tags.yaml"), &cfg.Tags); err != nil {
		return nil, fmt.Errorf("load tags configuration: %w", err)
	}

	return cfg, nil
}

// loadYAML opens a YAML file and decodes its contents into out.
func loadYAML(path string, out any) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open %s: %w", path, err)
	}
	defer file.Close()

	if err := yaml.NewDecoder(file).Decode(out); err != nil {
		return fmt.Errorf("decode %s: %w", path, err)
	}

	return nil
}

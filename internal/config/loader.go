package config

import (
	"fmt"

	"os"

	"gopkg.in/yaml.v3"
)

// loadPlant reads and decodes plant.yaml.
func loadPlant(path string) (PlantConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return PlantConfig{}, fmt.Errorf("open plant config: %w", err)
	}
	defer file.Close()

	var plant PlantConfig
	if err := yaml.NewDecoder(file).Decode(&plant); err != nil {
		return PlantConfig{}, fmt.Errorf("decode plant config: %w", err)
	}

	return plant, nil
}

// loadCollector reads and decodes collector.yaml.
func loadCollector(path string) (CollectorConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return CollectorConfig{}, fmt.Errorf("open collector config: %w", err)
	}
	defer file.Close()

	var collector CollectorConfig
	if err := yaml.NewDecoder(file).Decode(&collector); err != nil {
		return CollectorConfig{}, fmt.Errorf("decode collector config: %w", err)
	}

	return collector, nil
}

// loadAPI reads and decodes api.yaml.
func loadAPI(path string) (APIConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return APIConfig{}, fmt.Errorf("open api config: %w", err)
	}
	defer file.Close()

	var api APIConfig
	if err := yaml.NewDecoder(file).Decode(&api); err != nil {
		return APIConfig{}, fmt.Errorf("decode api config: %w", err)
	}

	return api, nil
}

// loadAggregator reads and decodes aggregation.yaml.
func loadAggregator(path string) (AggregatorConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return AggregatorConfig{}, fmt.Errorf("open aggregator config: %w", err)
	}
	defer file.Close()

	var aggregator AggregatorConfig
	if err := yaml.NewDecoder(file).Decode(&aggregator); err != nil {
		return AggregatorConfig{}, fmt.Errorf("decode aggregator config: %w", err)
	}

	return aggregator, nil
}

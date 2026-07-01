package config

import (
	"time"

	"pharma-platform/internal/models"
)

// Config represents the complete application configuration.
type Config struct {
	Plant      PlantConfig
	Collector  CollectorConfig
	API        APIConfig
	Aggregator AggregatorConfig

	PLCs []models.PLC
	Tags []models.Tag
}

// PlantConfig contains metadata about the manufacturing plant.
type PlantConfig struct {
	Name     string `yaml:"name"`
	Location string `yaml:"location"`
	TimeZone string `yaml:"timezone"`
}

// CollectorConfig contains runtime settings for the telemetry collector.
type CollectorConfig struct {
	Workers    int `yaml:"workers"`
	BufferSize int `yaml:"buffer_size"`
}

// APIConfig contains HTTP server configuration.
type APIConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

// AggregatorConfig contains runtime settings for the aggregation service.
type AggregatorConfig struct {
	Interval time.Duration `yaml:"interval"`
}

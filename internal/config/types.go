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
	Name     string
	Location string
	TimeZone string
}

// CollectorConfig contains runtime settings for the telemetry collector.
type CollectorConfig struct {
	Workers    int
	BufferSize int
}

// APIConfig contains HTTP server configuration.
type APIConfig struct {
	Host string
	Port int
}

// AggregatorConfig contains runtime settings for the aggregation service.
type AggregatorConfig struct {
	Interval time.Duration
}

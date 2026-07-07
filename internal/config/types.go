package config

import (
	"time"

	"pharma-platform/internal/postgres"
	"pharma-platform/internal/questdb"
)

type Config struct {
	Plant      PlantConfig      `yaml:"plant"`
	Collector  CollectorConfig  `yaml:"collector"`
	API        APIConfig        `yaml:"api"`
	Aggregator AggregatorConfig `yaml:"aggregator"`
	Postgres   postgres.Config  `yaml:"postgres"`
	QuestDB    questdb.Config   `yaml:"questdb"`
}

type PlantConfig struct {
	Name     string `yaml:"name"`
	Location string `yaml:"location"`
	TimeZone string `yaml:"timezone"`
}

type CollectorConfig struct {
	Workers   int `yaml:"workers"`
	QueueSize int `yaml:"queue_size"`
}

type APIConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type AggregatorConfig struct {
	Interval time.Duration `yaml:"interval"`
}

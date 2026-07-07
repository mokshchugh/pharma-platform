package config

import "fmt"

func Validate(cfg *Config) error {
	if cfg.Plant.Name == "" {
		return fmt.Errorf("plant name is required")
	}

	if cfg.Collector.Workers <= 0 {
		cfg.Collector.Workers = 16
	}
	if cfg.Collector.QueueSize <= 0 {
		cfg.Collector.QueueSize = 10000
	}

	if cfg.API.Host == "" {
		cfg.API.Host = "0.0.0.0"
	}
	if cfg.API.Port == 0 {
		cfg.API.Port = 8081
	}

	if cfg.Postgres.Host == "" {
		cfg.Postgres.Host = "localhost"
	}
	if cfg.Postgres.Port == 0 {
		cfg.Postgres.Port = 5432
	}
	if cfg.Postgres.Database == "" {
		cfg.Postgres.Database = "pharma"
	}
	if cfg.Postgres.User == "" {
		cfg.Postgres.User = "postgres"
	}

	if cfg.QuestDB.Host == "" {
		cfg.QuestDB.Host = "localhost"
	}
	if cfg.QuestDB.Port == 0 {
		cfg.QuestDB.Port = 9009
	}

	return nil
}

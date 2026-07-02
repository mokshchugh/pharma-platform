package main

import (
	"log"
	"os"
	"time"

	"gopkg.in/yaml.v3"

	"pharma-platform/internal/api"
	"pharma-platform/internal/api/handlers"
	"pharma-platform/internal/config"
	"pharma-platform/internal/questdb"
)

func main() {
	// cfg, err := config.Load("config")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// if err := config.Validate(cfg); err != nil {
	// 	log.Fatal(err)
	// }

	var cfg config.APIConfig

	data, err := os.ReadFile("config/api.yaml")
	if err != nil {
		log.Fatal(err)
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		log.Fatal(err)
	}

	client := questdb.New(
		questdb.Config{
			Host:          "localhost",
			Port:          9009,
			BatchSize:     1000,
			FlushInterval: time.Second,
		},
	)

	reader := questdb.NewReader(client)

	telemetry := handlers.NewTelemetryHandler(
		reader,
	)

	server := api.New(
		cfg,
		telemetry,
	)

	log.Printf(
		"API listening on %s:%d",
		cfg.Host,
		cfg.Port,
	)

	if err := server.Start(); err != nil {
		log.Fatal(err)
	}
}

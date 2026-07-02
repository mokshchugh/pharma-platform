package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"pharma-platform/internal/collector"
	"pharma-platform/internal/config"
	"pharma-platform/internal/models"
	"pharma-platform/internal/plc/drivers/opcua"
	"pharma-platform/internal/postgres"
	"pharma-platform/internal/questdb"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.Load("config")
	if err != nil {
		log.Fatal(err)
	}

	if err := config.Validate(cfg); err != nil {
		log.Fatal(err)
	}

	samples := make(chan models.Sample, 100000)

	plc := cfg.PLCs[0]

	driver := opcua.New(opcua.NewConfig(plc))

	collectorService := collector.New(
		driver,
		collector.Config{
			Workers:   8,
			QueueSize: 1000,
		},
		cfg.Tags,
		samples,
	)

	questClient := questdb.New(
		questdb.Config{
			Host:          "localhost",
			Port:          9009,
			BatchSize:     1000,
			FlushInterval: cfg.Aggregator.Interval,
		},
	)

	writer := questdb.NewWriter(
		questClient,
		samples,
	)

	postgresClient := postgres.New(
		postgres.Config{
			Host:         "localhost",
			Port:         5432,
			Database:     "pharma",
			User:         "postgres",
			Password:     "postgres",
			MaxOpenConns: 20,
			MaxIdleConns: 10,
		},
	)

	defer postgresClient.Close()

	if err := collectorService.Start(ctx); err != nil {
		log.Fatal(err)
	}

	go func() {
		if err := writer.Start(ctx); err != nil {
			log.Fatal(err)
		}
	}()

	log.Println("Pharma Platform started.")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	<-sig

	log.Println("Shutting down...")

	cancel()

	collectorService.Stop()
	writer.Stop()
}

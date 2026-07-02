package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"pharma-platform/internal/aggregator"
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

	driver := opcua.New(opcua.NewConfig(cfg.PLCs[0]))

	collectorService := collector.New(
		driver,
		cfg.Collector,
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

	if err := questClient.Connect(ctx); err != nil {
		log.Fatal(err)
	}

	writer := questdb.NewWriter(
		questClient,
		"plc_samples",
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

	if err := postgresClient.Connect(ctx); err != nil {
		log.Fatal(err)
	}
	defer postgresClient.Close()

	aggregatorService := aggregator.New(
		questClient,
		postgres.NewWriter(postgresClient),
		aggregator.Config{
			Interval: cfg.Aggregator.Interval,
		},
	)

	if err := collectorService.Start(ctx); err != nil {
		log.Fatal(err)
	}

	go func() {
		if err := writer.Start(ctx); err != nil {
			log.Fatal(err)
		}
	}()

	if err := aggregatorService.Start(ctx); err != nil {
		log.Fatal(err)
	}

	log.Println("Pharma Platform started.")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	<-sig

	log.Println("Shutting down...")

	cancel()

	collectorService.Stop()
	writer.Stop()
	aggregatorService.Stop()
}

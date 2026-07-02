package main

import (
	"context"
	"log"
	"math/rand/v2"
	"time"

	"pharma-platform/internal/models"
	"pharma-platform/internal/questdb"
)

const (
	SamplesPerTick = 100
	TickInterval   = time.Millisecond
)

func main() {
	ctx := context.Background()

	samples := make(chan models.Sample, 100000)

	client := questdb.New(
		questdb.Config{
			Host:          "localhost",
			Port:          9009,
			BatchSize:     1000,
			FlushInterval: time.Second,
		},
	)

	writer := questdb.NewWriter(
		client,
		"plc_samples",
		samples,
	)

	go func() {
		if err := writer.Start(ctx); err != nil {
			log.Fatal(err)
		}
	}()

	simulate(samples)
}

func simulate(samples chan<- models.Sample) {
	ticker := time.NewTicker(TickInterval)
	defer ticker.Stop()

	for range ticker.C {
		for i := 0; i < SamplesPerTick; i++ {
			samples <- models.Sample{
				Timestamp: time.Now(),
				PLCID:     "plc-1",
				TagID:     "motor_speed",
				Value:     rand.Float64() * 1000,
				Quality:   models.QualityGood,
			}
		}
	}
}

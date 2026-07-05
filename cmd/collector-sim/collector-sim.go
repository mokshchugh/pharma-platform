package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"pharma-platform/internal/collector"
	"pharma-platform/internal/config"
	"pharma-platform/internal/models"
	"pharma-platform/internal/plc"
	"pharma-platform/internal/questdb"
)

type MockDriver struct{}

var _ plc.Driver = (*MockDriver)(nil)

func (m *MockDriver) Connect(ctx context.Context) error {
	return nil
}

func (m *MockDriver) Close() error {
	return nil
}

func (m *MockDriver) Read(
	ctx context.Context,
	tag models.Tag,
) (models.Sample, error) {

	return models.Sample{
		Timestamp: time.Now(),
		PLCID:     tag.PLCID,
		TagID:     tag.ID,
		Value:     42.0,
		Quality:   models.QualityGood,
	}, nil
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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
			log.Printf("questdb writer stopped: %v", err)
		}
	}()

	cfg := config.CollectorConfig{
		Workers:   16,
		QueueSize: 10000,
	}

	tags := make([]models.Tag, 1000)

	for i := range tags {
		tags[i] = models.Tag{
			ID:           fmt.Sprintf("tag-%d", i),
			PLCID:        "plc-1",
			Name:         fmt.Sprintf("Tag %d", i),
			Address:      "mock",
			DataType:     models.DataTypeFloat64,
			PollInterval: 100 * time.Millisecond,
			Enabled:      true,
		}
	}

	c := collector.New(
		&MockDriver{},
		cfg,
		tags,
		samples,
	)

	if err := c.Start(ctx); err != nil {
		log.Fatal(err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Println("shutting down...")
	cancel()
	c.Stop()
	close(samples)
	client.Close()
}

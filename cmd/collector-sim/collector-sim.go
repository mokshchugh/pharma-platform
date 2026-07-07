package main

import (
	"context"
	"fmt"
	"log"
	"math"
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
	val := 42.0 + math.Sin(float64(time.Now().UnixMilli())/1000.0)*10.0
	return models.Sample{
		Timestamp: time.Now(),
		PLCID:     tag.PLCID,
		TagID:     tag.ID,
		Value:     val,
		Quality:   models.QualityGood,
	}, nil
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	plcNames := []struct {
		id   string
		name string
	}{
		{"plc-1", "Fluid Bed Dryer"},
		{"plc-2", "Tablet Press"},
		{"plc-3", "HVAC System"},
	}

	tags := make([]models.Tag, 0, len(plcNames)*25)
	for _, p := range plcNames {
		for i := range 25 {
			tagID := fmt.Sprintf("%s-tag-%02d", p.id, i)
			tags = append(tags, models.Tag{
				ID:           tagID,
				PLCID:        p.id,
				Name:         fmt.Sprintf("%s Tag %d", p.name, i),
				Address:      fmt.Sprintf("mock://%s/%d", p.id, i),
				DataType:     models.DataTypeFloat64,
				PollInterval: 100 * time.Millisecond,
				Enabled:      true,
			})
		}
	}

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

	c := collector.New(
		&MockDriver{},
		cfg,
		tags,
		samples,
	)

	if err := c.Start(ctx); err != nil {
		log.Fatal(err)
	}

	log.Println("collector-sim started · 3 PLCs · 75 tags · SIGUSR1=pause · SIGUSR2=resume · Ctrl+C=stop")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1, syscall.SIGUSR2)

	for {
		s := <-sigCh

		switch s {
		case syscall.SIGUSR1:
			c.Pause()
			log.Println("collector paused")

		case syscall.SIGUSR2:
			c.Resume()
			log.Println("collector resumed")

		default:
			log.Println("shutting down...")

			c.Stop()
			close(samples)
			writer.Stop()
			cancel()
			return
		}
	}
}

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
	"pharma-platform/internal/postgres"
	"pharma-platform/internal/questdb"
	"pharma-platform/internal/store"
)

type MockDriver struct{}

var _ plc.Driver = (*MockDriver)(nil)

func (m *MockDriver) Connect(ctx context.Context) error { return nil }
func (m *MockDriver) Close() error                      { return nil }
func (m *MockDriver) Read(ctx context.Context, tag models.Tag) (models.Sample, error) {
	base := 42.0
	switch tag.DataType {
	case models.DataTypeBool:
		base = 1.0
	case models.DataTypeInt16, models.DataTypeInt32:
		base = 100.0
	case models.DataTypeFloat32, models.DataTypeFloat64:
		base = 42.0
	}
	val := base + math.Sin(float64(time.Now().UnixMilli())/1000.0)*10.0
	return models.Sample{
		Timestamp:   time.Now(),
		MachineID:   fmt.Sprintf("%d", tag.MachineID),
		MachineName: tag.MachineName,
		TagName:     tag.Name,
		Value:       val,
		Quality:     models.QualityGood,
	}, nil
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.Load("config/bootstrap.yaml")
	if err != nil {
		log.Fatal(err)
	}

	if err := config.Validate(cfg); err != nil {
		log.Fatal(err)
	}

	postgresClient := postgres.New(cfg.Postgres)
	if err := postgresClient.Connect(ctx); err != nil {
		log.Fatal(err)
	}
	defer postgresClient.Close()

	tagStore := store.NewTagStore(postgresClient)
	tags := tagStore.GetTags()

	log.Printf("collector-sim loaded %d tags from PostgreSQL", len(tags))
	if len(tags) == 0 {
		log.Fatal("no tags found — seed the database first (go run cmd/seed/main.go)")
	}

	questClient := questdb.New(cfg.QuestDB)
	if err := questClient.Connect(ctx); err != nil {
		log.Fatal(err)
	}
	defer questClient.Close()

	if err := store.MigrateQuestDB(ctx, questClient, "deploy/questdb/init"); err != nil {
		log.Fatal(err)
	}

	samples := make(chan models.Sample, 100000)

	writer := questdb.NewWriter(questClient, "plc_samples", samples)
	go func() {
		if err := writer.Start(ctx); err != nil {
			log.Printf("questdb writer stopped: %v", err)
		}
	}()

	collectorCfg := config.CollectorConfig{
		Workers:   16,
		QueueSize: 10000,
	}

	c := collector.New(&MockDriver{}, collectorCfg, tags, samples)
	if err := c.Start(ctx); err != nil {
		log.Fatal(err)
	}

	addr := fmt.Sprintf("http://localhost:%d/", cfg.API.Port)
	log.Printf("collector-sim started · %d machines from DB · writing to QuestDB · %s · SIGUSR1=pause · SIGUSR2=resume · Ctrl+C=stop", len(tagStore.GetTags()), addr)

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

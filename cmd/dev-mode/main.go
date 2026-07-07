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

	"pharma-platform/internal/api"
	"pharma-platform/internal/api/handlers"
	"pharma-platform/internal/collector"
	"pharma-platform/internal/config"
	"pharma-platform/internal/models"
	"pharma-platform/internal/plc"
	"pharma-platform/internal/questdb"
)

type mockDriver struct {
	offset float64
}

var _ plc.Driver = (*mockDriver)(nil)

func (m *mockDriver) Connect(ctx context.Context) error { return nil }
func (m *mockDriver) Close() error                      { return nil }
func (m *mockDriver) Read(ctx context.Context, tag models.Tag) (models.Sample, error) {
	val := 42.0 + m.offset + math.Sin(float64(time.Now().UnixMilli())/1000.0)*10.0
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

	plcs := make([]models.PLC, 0, len(plcNames))
	tags := make([]models.Tag, 0, len(plcNames)*25)

	for _, p := range plcNames {
		plcs = append(plcs, models.PLC{
			ID: p.id, Name: p.name, Driver: "mock", IPAddress: "localhost", Port: 0, Enabled: true,
		})

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

	collectorCfg := config.CollectorConfig{
		Workers:   16,
		QueueSize: 10000,
	}

	apiCfg := config.APIConfig{
		Host: "0.0.0.0",
		Port: 8081,
	}

	samples := make(chan models.Sample, 100000)

	driver := &mockDriver{}
	collectorService := collector.New(driver, collectorCfg, tags, samples)

	questClient := questdb.New(questdb.Config{
		Host:          "localhost",
		Port:          9009,
		BatchSize:     1000,
		FlushInterval: time.Second,
	})

	if err := questClient.Connect(ctx); err != nil {
		log.Fatal(err)
	}

	writer := questdb.NewWriter(questClient, "plc_samples", samples)

	if err := collectorService.Start(ctx); err != nil {
		log.Fatal(err)
	}

	go func() {
		if err := writer.Start(ctx); err != nil {
			log.Fatal(err)
		}
	}()

	reader := questdb.NewReader(questClient)
	plcStore := handlers.NewPLCConfigStore(plcs, tags)
	alarmStore := handlers.NewAlarmStore()

	telemetryHandler := handlers.NewTelemetryHandler(reader)
	plcHandler := handlers.NewPLCHandler(plcStore)
	tagHandler := handlers.NewTagHandler(plcStore)
	collectorHandler := handlers.NewCollectorHandler(collectorService)
	alarmHandler := handlers.NewAlarmHandler(alarmStore)
	systemHandler := handlers.NewSystemHandler(plcStore, alarmStore, collectorService)

	server := api.NewFull(apiCfg, &api.Handlers{
		Telemetry: telemetryHandler,
		PLC:       plcHandler,
		Tag:       tagHandler,
		Collector: collectorHandler,
		Alarms:    alarmHandler,
		System:    systemHandler,
	})

	go func() {
		log.Printf("dev-mode API listening on %s:%d", apiCfg.Host, apiCfg.Port)
		if err := server.Start(); err != nil {
			log.Fatal(err)
		}
	}()

	log.Println("dev-mode started · http://localhost:8081/ · SIGUSR1=pause · SIGUSR2=resume")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM, syscall.SIGUSR1, syscall.SIGUSR2)

	for {
		s := <-sig

		switch s {
		case syscall.SIGUSR1:
			collectorService.Pause()
			log.Println("collector paused")

		case syscall.SIGUSR2:
			collectorService.Resume()
			log.Println("collector resumed")

		default:
			log.Println("shutting down...")

			collectorService.Stop()
			close(samples)
			writer.Stop()
			cancel()

			_ = server.Stop(context.Background())
			return
		}
	}
}

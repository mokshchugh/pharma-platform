package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"pharma-platform/internal/api"
	"pharma-platform/internal/api/handlers"
	"pharma-platform/internal/config"
	"pharma-platform/internal/models"
	"pharma-platform/internal/questdb"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var apiCfg config.APIConfig
	apiCfg.Host = "0.0.0.0"
	apiCfg.Port = 8081

	plcs := []models.PLC{
		{ID: "plc-1", Name: "Fluid Bed Dryer", Driver: "opcua", IPAddress: "localhost", Port: 4840, Enabled: true},
		{ID: "plc-2", Name: "Tablet Press", Driver: "opcua", IPAddress: "localhost", Port: 4841, Enabled: true},
		{ID: "plc-3", Name: "HVAC System", Driver: "opcua", IPAddress: "localhost", Port: 4842, Enabled: true},
	}

	tags := make([]models.Tag, 0, 75)
	for _, p := range plcs {
		for i := range 25 {
			tags = append(tags, models.Tag{
				ID:           fmt.Sprintf("%s-tag-%02d", p.ID, i),
				PLCID:        p.ID,
				Name:         fmt.Sprintf("%s Tag %d", p.Name, i),
				Address:      fmt.Sprintf("ns=2;s=%s-tag-%02d", p.ID, i),
				DataType:     models.DataTypeFloat64,
				PollInterval: 100 * time.Millisecond,
				Enabled:      true,
			})
		}
	}

	client := questdb.New(questdb.Config{
		Host:          "localhost",
		Port:          9009,
		BatchSize:     1000,
		FlushInterval: 0,
	})

	reader := questdb.NewReader(client)
	plcStore := handlers.NewPLCConfigStore(plcs, tags)
	alarmStore := handlers.NewAlarmStore()

	telemetryHandler := handlers.NewTelemetryHandler(reader)
	plcHandler := handlers.NewPLCHandler(plcStore)
	tagHandler := handlers.NewTagHandler(plcStore)

	dummyCollector := &dummyCollector{}
	collectorHandler := handlers.NewCollectorHandler(dummyCollector)
	alarmHandler := handlers.NewAlarmHandler(alarmStore)
	systemHandler := handlers.NewSystemHandler(plcStore, alarmStore, dummyCollector)

	server := api.NewFull(apiCfg, &api.Handlers{
		Telemetry: telemetryHandler,
		PLC:       plcHandler,
		Tag:       tagHandler,
		Collector: collectorHandler,
		Alarms:    alarmHandler,
		System:    systemHandler,
	})

	go func() {
		log.Printf("API listening on %s:%d", apiCfg.Host, apiCfg.Port)
		if err := server.Start(); err != nil {
			log.Fatal(err)
		}
	}()

	log.Println("standalone API started · http://localhost:8081/")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	<-sig
	log.Println("shutting down...")

	_ = server.Stop(ctx)
	cancel()
}

type dummyCollector struct{}

func (d *dummyCollector) IsPaused() bool    { return false }
func (d *dummyCollector) Pause()             {}
func (d *dummyCollector) Resume()            {}
func (d *dummyCollector) TickCount() int64   { return 0 }
func (d *dummyCollector) DispatchSum() int64 { return 0 }

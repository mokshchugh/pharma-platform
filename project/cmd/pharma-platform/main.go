package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"pharma-platform/internal/api"
	"pharma-platform/internal/api/handlers"
	"pharma-platform/internal/config"
	"pharma-platform/internal/postgres"
	"pharma-platform/internal/questdb"
	"pharma-platform/internal/store"
)

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

	if err := store.MigratePostgres(ctx, postgresClient,
		"deploy/postgres/init",
		"",
		false,
	); err != nil {
		log.Fatal(err)
	}

	questClient := questdb.New(cfg.QuestDB)
	if err := questClient.Connect(ctx); err != nil {
		log.Fatal(err)
	}
	defer questClient.Close()

	if err := store.MigrateQuestDB(ctx, questClient, "deploy/questdb/init"); err != nil {
		log.Fatal(err)
	}

	machineStore := store.NewMachineStore(postgresClient)
	tagStore := store.NewTagStore(postgresClient)
	reader := questdb.NewReader(questClient)
	alarmStore := handlers.NewAlarmStore()
	dummyCollector := &dummyCollector{}

	telemetryHandler := handlers.NewTelemetryHandler(reader)
	plcHandler := handlers.NewPLCHandler(machineStore)
	tagHandler := handlers.NewTagHandler(tagStore)
	collectorHandler := handlers.NewCollectorHandler(dummyCollector)
	alarmHandler := handlers.NewAlarmHandler(alarmStore)
	systemHandler := handlers.NewSystemHandler(machineStore, alarmStore, dummyCollector)

	server := api.NewFull(cfg.API, &api.Handlers{
		Telemetry: telemetryHandler,
		PLC:       plcHandler,
		Tag:       tagHandler,
		Collector: collectorHandler,
		Alarms:    alarmHandler,
		System:    systemHandler,
	})

	go func() {
		log.Printf("Pharma Platform listening on %s:%d", cfg.API.Host, cfg.API.Port)
		if err := server.Start(); err != nil {
			log.Fatal(err)
		}
	}()

	log.Println("Pharma Platform started (production mode)")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	<-sig
	log.Println("Shutting down...")

	_ = server.Stop(ctx)
	cancel()
}

type dummyCollector struct{}

func (d *dummyCollector) IsPaused() bool    { return false }
func (d *dummyCollector) Pause()             {}
func (d *dummyCollector) Resume()            {}
func (d *dummyCollector) TickCount() int64   { return 0 }
func (d *dummyCollector) DispatchSum() int64 { return 0 }

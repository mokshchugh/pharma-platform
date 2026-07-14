package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"pharma-platform/internal/api"
	"pharma-platform/internal/api/handlers"
	"pharma-platform/internal/business"
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
	productionStore := store.NewProductionStore(postgresClient)
	reader := questdb.NewReader(questClient)
	alarmStore := handlers.NewAlarmStore()
	dummyCollector := &dummyCollector{}

	telemetryHandler := handlers.NewTelemetryHandler(reader)
	plcHandler := handlers.NewPLCHandler(machineStore)
	tagHandler := handlers.NewTagHandler(tagStore)
	machineHandler := handlers.NewMachineHandler(machineStore, reader)
	analyticsHandler := handlers.NewAnalyticsHandler(tagStore, reader)
	collectorHandler := handlers.NewCollectorHandler(dummyCollector)
	alarmHandler := handlers.NewAlarmHandler(alarmStore)
	systemHandler := handlers.NewSystemHandler(machineStore, alarmStore, dummyCollector)
	dashboardHandler := handlers.NewDashboardHandler(productionStore, alarmStore)
	oeeHandler := handlers.NewOEEHandler(productionStore)
	productionHandler := handlers.NewProductionHandler(productionStore)
	controlHandler := handlers.NewControlHandler()

	bizEngine := business.NewEngine(business.SimulatorConfig{
		MachineCount:    len(machineStore.GetPLCs()),
		AlarmStore:      alarmStore,
		CollectorPaused: dummyCollector.IsPaused,
	})
	go func() {
		tick := time.NewTicker(5 * time.Second)
		defer tick.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-tick.C:
				bizEngine.Tick()
			}
		}
	}()

	bizAnalyticsHandler := handlers.NewBusinessAnalyticsHandler(bizEngine)

	server := api.NewFull(cfg.API, &api.Handlers{
		Telemetry:    telemetryHandler,
		PLC:          plcHandler,
		Tag:          tagHandler,
		Machine:      machineHandler,
		Analytics:    analyticsHandler,
		BizAnalytics: bizAnalyticsHandler,
		Collector:    collectorHandler,
		Alarms:       alarmHandler,
		System:       systemHandler,
		Dashboard:    dashboardHandler,
		OEE:          oeeHandler,
		Production:   productionHandler,
		Controls:     controlHandler,
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

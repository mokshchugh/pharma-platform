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
	"pharma-platform/internal/business"
	"pharma-platform/internal/config"
	"pharma-platform/internal/models"
	"pharma-platform/internal/postgres"
	"pharma-platform/internal/questdb"
	"pharma-platform/internal/simulation"
	"pharma-platform/internal/store"
	"pharma-platform/internal/telemetry"
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
		"deploy/postgres/seed",
		true,
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
	alarmAckStore := store.NewAlarmAckStore(postgresClient)
	controlStore := store.NewControlStore(postgresClient)
	productionStore.CloseStaleRunsAndDowntime()

	plcs := machineStore.GetPLCs()
	tags := tagStore.GetTags()

	log.Printf("dev-mode loaded %d machines and %d tags from PostgreSQL", len(plcs), len(tags))

	if len(tags) == 0 {
		log.Fatal("no tags found in database — run seed first (go run cmd/seed/main.go)")
	}

	allMachines, _ := machineStore.GetAllMachines()
	machineIDs := make([]int, 0, len(allMachines))
	for _, m := range allMachines {
		machineIDs = append(machineIDs, m.ID)
	}
	controlStore.EnsureDefaults(machineIDs)

	reader := questdb.NewReader(questClient)
	alarmStore := handlers.NewAlarmStore(reader, alarmAckStore)

	rawSamples := make(chan models.Sample, 100000)
	writerSamples := make(chan models.Sample, 100000)

	writer := questdb.NewWriter(questClient, "plc_samples", writerSamples)

	go func() {
		if err := writer.Start(ctx); err != nil {
			log.Fatal(err)
		}
	}()

	// Tee raw samples to the writer while deriving alarms/run-state/
	// production-runs/downtime on the way through via the shared
	// telemetry.Tracker — the same derivation a real driver-fed pipeline
	// would use, so simulated runs exercise the real code path.
	tracker := telemetry.NewTracker(reader, productionStore)
	go func() {
		for s := range rawSamples {
			tracker.Observe(ctx, s)

			select {
			case writerSamples <- s:
			case <-ctx.Done():
				close(writerSamples)
				return
			}
		}
		close(writerSamples)
	}()

	sim := simulation.New(rawSamples)
	go func() {
		tick := time.NewTicker(100 * time.Millisecond)
		defer tick.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-tick.C:
				sim.Tick()
			}
		}
	}()

	telemetryHandler := handlers.NewTelemetryHandler(reader)
	plcHandler := handlers.NewPLCHandler(machineStore)
	tagHandler := handlers.NewTagHandler(tagStore)
	machineHandler := handlers.NewMachineHandler(machineStore, reader)
	analyticsHandler := handlers.NewAnalyticsHandler(tagStore, reader)
	collectorHandler := handlers.NewCollectorHandler(sim)
	alarmHandler := handlers.NewAlarmHandler(alarmStore)
	systemHandler := handlers.NewSystemHandler(machineStore, alarmStore, sim)
	dashboardHandler := handlers.NewDashboardHandler(productionStore, alarmStore)
	oeeHandler := handlers.NewOEEHandler(productionStore)
	productionHandler := handlers.NewProductionHandler(productionStore)
	controlHandler := handlers.NewControlHandler(controlStore)

	bizEngine := business.NewEngine(business.RealEngineConfig{
		ProductionStore: productionStore,
		MachineStore:    machineStore,
		TagStore:        tagStore,
		Reader:          reader,
		AlarmAckStore:   alarmAckStore,
		CollectorPaused: sim.IsPaused,
	})

	bizAnalyticsHandler := handlers.NewBusinessAnalyticsHandler(bizEngine)

	server := api.NewBackend(cfg.API, &api.Handlers{
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
		log.Printf("dev-mode API listening on %s:%d", cfg.API.Host, cfg.API.Port)
		if err := server.Start(); err != nil {
			log.Fatal(err)
		}
	}()

	addr := fmt.Sprintf("http://localhost:%d/", cfg.API.Port)
	log.Printf("dev-mode started · %s · SIGUSR1=pause · SIGUSR2=resume", addr)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM, syscall.SIGUSR1, syscall.SIGUSR2)

	for {
		s := <-sig
		switch s {
		case syscall.SIGUSR1:
			sim.Pause()
			log.Println("simulator paused")
		case syscall.SIGUSR2:
			sim.Resume()
			log.Println("simulator resumed")
		default:
			log.Println("shutting down...")
			close(rawSamples)
			writer.Stop()
			cancel()
			_ = server.Stop(context.Background())
			return
		}
	}
}

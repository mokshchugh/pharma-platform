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
	"pharma-platform/internal/collector"
	"pharma-platform/internal/config"
	"pharma-platform/internal/models"
	"pharma-platform/internal/plc"
	"pharma-platform/internal/postgres"
	"pharma-platform/internal/questdb"
	"pharma-platform/internal/simulation"
	"pharma-platform/internal/store"
)

type mockDriver struct {
	offset float64
}

var _ plc.Driver = (*mockDriver)(nil)

func (m *mockDriver) Connect(ctx context.Context) error { return nil }
func (m *mockDriver) Close() error                      { return nil }
func (m *mockDriver) Read(ctx context.Context, tag models.Tag) (models.Sample, error) {
	base := 42.0
	switch tag.DataType {
	case models.DataTypeBool:
		base = 1.0
	case models.DataTypeInt16, models.DataTypeInt32:
		base = 100.0
	case models.DataTypeFloat32, models.DataTypeFloat64:
		base = 42.0
	}
	val := base + m.offset // simple offset, real data comes from simulation
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
	alarmStore := handlers.NewAlarmStore()

	plcs := machineStore.GetPLCs()
	tags := tagStore.GetTags()

	log.Printf("dev-mode loaded %d machines and %d tags from PostgreSQL", len(plcs), len(tags))

	if len(tags) == 0 {
		log.Fatal("no tags found in database — run seed first (go run cmd/seed/main.go)")
	}

	samples := make(chan models.Sample, 100000)

	collectorCfg := config.CollectorConfig{
		Workers:   16,
		QueueSize: 10000,
	}

	driver := &mockDriver{}
	collectorService := collector.New(driver, collectorCfg, tags, samples)

	writer := questdb.NewWriter(questClient, "plc_samples", samples)

	if err := collectorService.Start(ctx); err != nil {
		log.Fatal(err)
	}

	go func() {
		if err := writer.Start(ctx); err != nil {
			log.Fatal(err)
		}
	}()

	sim := simulation.New(samples)
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

	reader := questdb.NewReader(questClient)

	telemetryHandler := handlers.NewTelemetryHandler(reader)
	plcHandler := handlers.NewPLCHandler(machineStore)
	tagHandler := handlers.NewTagHandler(tagStore)
	machineHandler := handlers.NewMachineHandler(machineStore, reader)
	analyticsHandler := handlers.NewAnalyticsHandler(tagStore, reader)
	collectorAdapter := &handlers.CollectorAdapter{C: collectorService}
	collectorHandler := handlers.NewCollectorHandler(collectorAdapter)
	alarmHandler := handlers.NewAlarmHandler(alarmStore)
	systemHandler := handlers.NewSystemHandler(machineStore, alarmStore, collectorService)
	dashboardHandler := handlers.NewDashboardHandler(productionStore, alarmStore)
	oeeHandler := handlers.NewOEEHandler(productionStore)
	productionHandler := handlers.NewProductionHandler(productionStore)
	controlHandler := handlers.NewControlHandler()

	bizEngine := business.NewEngine(business.SimulatorConfig{
		MachineCount:    len(plcs),
		AlarmStore:      alarmStore,
		CollectorPaused: collectorService.IsPaused,
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

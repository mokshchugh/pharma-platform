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

	// Tee raw samples to the writer while evaluating alarm/run-state
	// transitions on the way through, since nothing else in the pipeline
	// evaluates tag thresholds. This is what populates the QuestDB `alarms`
	// and `machine_state` tables that the real alarm store and MTBF/MTTR
	// analysis read from, and drives the Postgres `production_runs` /
	// `downtime_events` rows that OEE and the Production page read from.
	go func() {
		lastAlarmActive := make(map[string]bool)
		lastRunning := make(map[string]bool)
		activeRun := make(map[string]int)
		activeDowntime := make(map[string]int)
		lastGood := make(map[string]int)
		lastBad := make(map[string]int)

		machineIDFrom := func(s string) int {
			id := 0
			fmt.Sscanf(s, "%d", &id)
			return id
		}

		toFloat := func(v any) float64 {
			switch n := v.(type) {
			case float64:
				return n
			case float32:
				return float64(n)
			case int:
				return float64(n)
			case int32:
				return float64(n)
			case int64:
				return float64(n)
			default:
				return 0
			}
		}

		for s := range rawSamples {
			switch s.TagName {
			case "Alarm_Status", "AlarmStatus":
				active := toFloat(s.Value) == 1
				if active && !lastAlarmActive[s.MachineID] {
					severity := "warning"
					if fault, ok := lastRunning[s.MachineID]; ok && !fault {
						severity = "critical"
					}
					msg := fmt.Sprintf("%s alarm active on machine %s", s.TagName, s.MachineID)
					if err := reader.InsertAlarm(ctx, s.MachineID, s.TagName, severity, msg); err != nil {
						log.Printf("insert alarm: %v", err)
					}
				}
				lastAlarmActive[s.MachineID] = active

			case "Run_Status", "RunStatus":
				running := toFloat(s.Value) == 1
				if prev, ok := lastRunning[s.MachineID]; !ok || prev != running {
					state := "stopped"
					if running {
						state = "running"
					}
					if err := reader.InsertMachineState(ctx, s.MachineID, state, 0, 0, 0); err != nil {
						log.Printf("insert machine_state: %v", err)
					}

					machineID := machineIDFrom(s.MachineID)
					if running {
						if downID, ok := activeDowntime[s.MachineID]; ok {
							_ = productionStore.EndDowntime(downID)
							delete(activeDowntime, s.MachineID)
						}
						if _, ok := activeRun[s.MachineID]; !ok {
							if run, err := productionStore.CreateRun(machineID, fmt.Sprintf("SIM-%s-%d", s.MachineID, s.Timestamp.Unix()), "", 0); err == nil {
								activeRun[s.MachineID] = run.ID
							}
						}
					} else {
						if runID, ok := activeRun[s.MachineID]; ok {
							_ = productionStore.UpdateRunCounts(runID, lastGood[s.MachineID], lastBad[s.MachineID])
							_ = productionStore.CompleteRun(runID)
							delete(activeRun, s.MachineID)
						}
						if _, ok := activeDowntime[s.MachineID]; !ok {
							if down, err := productionStore.StartDowntime(machineID, "machine stopped", "unplanned"); err == nil {
								activeDowntime[s.MachineID] = down.ID
							}
						}
					}
				}
				lastRunning[s.MachineID] = running

			case "Good_Count", "GoodCount", "Good_Print_Count":
				lastGood[s.MachineID] = int(toFloat(s.Value))
				if runID, ok := activeRun[s.MachineID]; ok {
					_ = productionStore.UpdateRunCounts(runID, lastGood[s.MachineID], lastBad[s.MachineID])
				}

			case "Reject_Count", "RejectCount":
				lastBad[s.MachineID] = int(toFloat(s.Value))
				if runID, ok := activeRun[s.MachineID]; ok {
					_ = productionStore.UpdateRunCounts(runID, lastGood[s.MachineID], lastBad[s.MachineID])
				}
			}

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

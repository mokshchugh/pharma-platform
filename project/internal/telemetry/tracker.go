// Package telemetry derives higher-level state (alarms, machine
// running/stopped state, production runs, downtime events) from a raw
// stream of tag samples. It doesn't care whether the samples came from
// the built-in simulator or a real PLC driver — any producer that feeds
// samples through Observe gets the same alarms/OEE/production behavior,
// so simulated runs are a faithful stand-in for how a real deployment's
// data would flow through the same pipeline.
package telemetry

import (
	"context"
	"fmt"
	"log"

	"pharma-platform/internal/models"
	"pharma-platform/internal/questdb"
	"pharma-platform/internal/store"
)

type Tracker struct {
	reader          *questdb.Reader
	productionStore *store.ProductionStore

	lastAlarmActive map[string]bool
	lastRunning     map[string]bool
	activeRun       map[string]int
	activeDowntime  map[string]int

	lastGood map[string]int
	lastBad  map[string]int
	baseGood map[string]int // lifetime counter value when the active run started
	baseBad  map[string]int
}

func NewTracker(reader *questdb.Reader, productionStore *store.ProductionStore) *Tracker {
	return &Tracker{
		reader:          reader,
		productionStore: productionStore,
		lastAlarmActive: make(map[string]bool),
		lastRunning:     make(map[string]bool),
		activeRun:       make(map[string]int),
		activeDowntime:  make(map[string]int),
		lastGood:        make(map[string]int),
		lastBad:         make(map[string]int),
		baseGood:        make(map[string]int),
		baseBad:         make(map[string]int),
	}
}

func machineIDFrom(s string) int {
	id := 0
	fmt.Sscanf(s, "%d", &id)
	return id
}

func toFloat(v any) float64 {
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

// Observe processes one raw sample, updating alarms/machine_state in
// QuestDB and production_runs/downtime_events in Postgres as needed.
func (t *Tracker) Observe(ctx context.Context, s models.Sample) {
	switch s.TagName {
	case "Alarm_Status", "AlarmStatus":
		t.observeAlarm(ctx, s)
	case "Run_Status", "RunStatus":
		t.observeRunStatus(ctx, s)
	case "Good_Count", "GoodCount", "Good_Print_Count":
		t.lastGood[s.MachineID] = int(toFloat(s.Value))
		t.flushRunCounts(s.MachineID)
	case "Reject_Count", "RejectCount":
		t.lastBad[s.MachineID] = int(toFloat(s.Value))
		t.flushRunCounts(s.MachineID)
	}
}

func (t *Tracker) observeAlarm(ctx context.Context, s models.Sample) {
	active := toFloat(s.Value) == 1
	if active && !t.lastAlarmActive[s.MachineID] {
		severity := "warning"
		if fault, ok := t.lastRunning[s.MachineID]; ok && !fault {
			severity = "critical"
		}
		msg := fmt.Sprintf("%s alarm active on machine %s", s.TagName, s.MachineID)
		if err := t.reader.InsertAlarm(ctx, s.MachineID, s.TagName, severity, msg); err != nil {
			log.Printf("insert alarm: %v", err)
		}
	}
	t.lastAlarmActive[s.MachineID] = active
}

func (t *Tracker) observeRunStatus(ctx context.Context, s models.Sample) {
	running := toFloat(s.Value) == 1
	prev, ok := t.lastRunning[s.MachineID]
	if ok && prev == running {
		return
	}

	state := "stopped"
	if running {
		state = "running"
	}
	if err := t.reader.InsertMachineState(ctx, s.MachineID, state, 0, 0, 0); err != nil {
		log.Printf("insert machine_state: %v", err)
	}

	machineID := machineIDFrom(s.MachineID)
	if running {
		if downID, ok := t.activeDowntime[s.MachineID]; ok {
			_ = t.productionStore.EndDowntime(downID)
			delete(t.activeDowntime, s.MachineID)
		}
		if _, ok := t.activeRun[s.MachineID]; !ok {
			run, err := t.productionStore.CreateRun(machineID, fmt.Sprintf("SIM-%s-%d", s.MachineID, s.Timestamp.Unix()), "", 0)
			if err == nil {
				t.activeRun[s.MachineID] = run.ID
				// Lifetime counters keep their value across stops, so a
				// run's good/bad counts must be measured as a delta from
				// this baseline, not the raw counter value.
				t.baseGood[s.MachineID] = t.lastGood[s.MachineID]
				t.baseBad[s.MachineID] = t.lastBad[s.MachineID]
			}
		}
	} else {
		if runID, ok := t.activeRun[s.MachineID]; ok {
			t.flushRunCounts(s.MachineID)
			_ = t.productionStore.CompleteRun(runID)
			delete(t.activeRun, s.MachineID)
		}
		if _, ok := t.activeDowntime[s.MachineID]; !ok {
			down, err := t.productionStore.StartDowntime(machineID, "machine stopped", "unplanned")
			if err == nil {
				t.activeDowntime[s.MachineID] = down.ID
			}
		}
	}

	t.lastRunning[s.MachineID] = running
}

func (t *Tracker) flushRunCounts(machineID string) {
	runID, ok := t.activeRun[machineID]
	if !ok {
		return
	}
	good := t.lastGood[machineID] - t.baseGood[machineID]
	bad := t.lastBad[machineID] - t.baseBad[machineID]
	if good < 0 {
		good = 0
	}
	if bad < 0 {
		bad = 0
	}
	_ = t.productionStore.UpdateRunCounts(runID, good, bad)
}

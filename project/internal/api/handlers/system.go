package handlers

import (
	"encoding/json"
	"net/http"
)

type SystemHandler struct {
	plcStore   PLCStore
	alarmStore *AlarmStore
	collector  CollectorStatusProvider
}

type CollectorStatusProvider interface {
	IsPaused() bool
	Pause()
	Resume()
}

func NewSystemHandler(
	plcStore PLCStore,
	alarmStore *AlarmStore,
	collector CollectorStatusProvider,
) *SystemHandler {
	return &SystemHandler{
		plcStore:   plcStore,
		alarmStore: alarmStore,
		collector:  collector,
	}
}

func (h *SystemHandler) Status(w http.ResponseWriter, r *http.Request) {
	plcs := h.plcStore.GetPLCs()
	online := 0
	for _, p := range plcs {
		if p.Enabled {
			online++
		}
	}

	activeAlarms := h.alarmStore.ActiveCount()
	criticalAlarms := h.alarmStore.CriticalCount()

	overallStatus := "ok"
	if criticalAlarms > 0 {
		overallStatus = "critical"
	} else if activeAlarms > 0 {
		overallStatus = "needs_attention"
	}

	collectorStatus := "running"
	if h.collector.IsPaused() {
		collectorStatus = "paused"
	}

	resp := map[string]any{
		"status": overallStatus,
		"plcs": map[string]int{
			"total":   len(plcs),
			"online":  online,
			"offline": len(plcs) - online,
		},
		"alarms": map[string]int{
			"active":   activeAlarms,
			"critical": criticalAlarms,
		},
		"collector": map[string]string{
			"status": collectorStatus,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

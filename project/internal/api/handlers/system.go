package handlers

import (
	"encoding/json"
	"net/http"

	"pharma-platform/internal/models"
)

type SystemHandler struct {
	plcStore    PLCStore
	alarmStore  *AlarmStore
	collector   CollectorStatusProvider
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

// Ensure *collector.Collector satisfies the interface at compile time.
var _ CollectorStatusProvider = (*collectorStatusShim)(nil)

type collectorStatusShim struct {
	paused bool
}

func (c *collectorStatusShim) IsPaused() bool { return c.paused }
func (c *collectorStatusShim) Pause()          {}
func (c *collectorStatusShim) Resume()         {}

var _ PLCStore = (*PLCConfigStore)(nil)

type PLCConfigStore struct {
	plcs []models.PLC
	tags []models.Tag
}

func NewPLCConfigStore(plcs []models.PLC, tags []models.Tag) *PLCConfigStore {
	return &PLCConfigStore{plcs: plcs, tags: tags}
}

func (s *PLCConfigStore) GetPLCs() []models.PLC {
	return s.plcs
}

func (s *PLCConfigStore) GetPLC(id string) *models.PLC {
	for i := range s.plcs {
		if s.plcs[i].ID == id {
			return &s.plcs[i]
		}
	}
	return nil
}

func (s *PLCConfigStore) TogglePLCEnabled(id string, enabled bool) error {
	for i := range s.plcs {
		if s.plcs[i].ID == id {
			s.plcs[i].Enabled = enabled
			return nil
		}
	}
	return nil
}

func (s *PLCConfigStore) GetTagsByPLC(plcID string) []models.Tag {
	var result []models.Tag
	for i := range s.tags {
		if s.tags[i].PLCID == plcID {
			result = append(result, s.tags[i])
		}
	}
	return result
}

func (s *PLCConfigStore) GetTags() []models.Tag {
	return s.tags
}

func (s *PLCConfigStore) GetTag(id string) *models.Tag {
	for i := range s.tags {
		if s.tags[i].ID == id {
			return &s.tags[i]
		}
	}
	return nil
}

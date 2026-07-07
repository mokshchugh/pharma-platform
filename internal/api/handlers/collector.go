package handlers

import (
	"encoding/json"
	"net/http"

	"pharma-platform/internal/collector"
)

type CollectorHandler struct {
	collector *collector.Collector
}

func NewCollectorHandler(c *collector.Collector) *CollectorHandler {
	return &CollectorHandler{collector: c}
}

func (h *CollectorHandler) Status(w http.ResponseWriter, r *http.Request) {
	status := "running"
	if h.collector.IsPaused() {
		status = "paused"
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":       status,
		"paused":       h.collector.IsPaused(),
		"tick_count":   h.collector.TickCount,
		"dispatch_sum": h.collector.DispatchSum,
	})
}

func (h *CollectorHandler) Pause(w http.ResponseWriter, r *http.Request) {
	h.collector.Pause()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "paused"})
}

func (h *CollectorHandler) Resume(w http.ResponseWriter, r *http.Request) {
	h.collector.Resume()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "resumed"})
}

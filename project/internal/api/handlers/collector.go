package handlers

import (
	"encoding/json"
	"net/http"
)

type CollectorHandle interface {
	IsPaused() bool
	Pause()
	Resume()
	TickCount() int64
	DispatchSum() int64
}

type CollectorHandler struct {
	handle CollectorHandle
}

func NewCollectorHandler(handle CollectorHandle) *CollectorHandler {
	return &CollectorHandler{handle: handle}
}

func (h *CollectorHandler) Status(w http.ResponseWriter, r *http.Request) {
	status := "running"
	if h.handle.IsPaused() {
		status = "paused"
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":       status,
		"paused":       h.handle.IsPaused(),
		"tick_count":   h.handle.TickCount(),
		"dispatch_sum": h.handle.DispatchSum(),
	})
}

func (h *CollectorHandler) Pause(w http.ResponseWriter, r *http.Request) {
	h.handle.Pause()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "paused"})
}

func (h *CollectorHandler) Resume(w http.ResponseWriter, r *http.Request) {
	h.handle.Resume()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "resumed"})
}

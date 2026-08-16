package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"pharma-platform/internal/store"

	"github.com/go-chi/chi/v5"
)

type ControlHandler struct {
	store *store.ControlStore
}

func NewControlHandler(controlStore *store.ControlStore) *ControlHandler {
	return &ControlHandler{store: controlStore}
}

// GET /api/v1/controls - list all machine control states
func (h *ControlHandler) List(w http.ResponseWriter, r *http.Request) {
	resp := h.store.List()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GET /api/v1/controls/{id}
func (h *ControlHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	state := h.store.Get(id)
	if state == nil {
		http.Error(w, "machine not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(state)
}

// POST /api/v1/controls/{id}/start
func (h *ControlHandler) Start(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.Atoi(idStr)
	h.store.Start(id)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "started"})
}

// POST /api/v1/controls/{id}/stop
func (h *ControlHandler) Stop(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.Atoi(idStr)
	h.store.Stop(id)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "stopped"})
}

// POST /api/v1/controls/{id}/setpoint
func (h *ControlHandler) Setpoint(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.Atoi(idStr)
	var req struct {
		Value float64 `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	h.store.SetSetpoint(id, req.Value)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "setpoint_updated"})
}

// POST /api/v1/controls/{id}/mode
func (h *ControlHandler) SetMode(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.Atoi(idStr)
	var req struct {
		Mode string `json:"mode"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	h.store.SetMode(id, req.Mode)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "mode_updated"})
}

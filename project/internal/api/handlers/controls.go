package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"

	"pharma-platform/internal/models"

	"github.com/go-chi/chi/v5"
)

type ControlHandler struct {
	mu       sync.RWMutex
	machines map[int]*models.MachineControlState
}

func NewControlHandler() *ControlHandler {
	states := make(map[int]*models.MachineControlState)
	for i := 1; i <= 11; i++ {
		states[i] = &models.MachineControlState{
			MachineID:   i,
			Running:     false,
			Speed:       0,
			Setpoint:    100,
			Mode:        "auto",
			Temperature: 25,
		}
	}
	return &ControlHandler{machines: states}
}

// GET /api/v1/controls - list all machine control states
func (h *ControlHandler) List(w http.ResponseWriter, r *http.Request) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	resp := make([]*models.MachineControlState, 0, len(h.machines))
	for _, s := range h.machines {
		resp = append(resp, s)
	}
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
	h.mu.RLock()
	state, ok := h.machines[id]
	h.mu.RUnlock()
	if !ok {
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
	h.mu.Lock()
	if s, ok := h.machines[id]; ok {
		s.Running = true
		s.Speed = s.Setpoint
	}
	h.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "started"})
}

// POST /api/v1/controls/{id}/stop
func (h *ControlHandler) Stop(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.Atoi(idStr)
	h.mu.Lock()
	if s, ok := h.machines[id]; ok {
		s.Running = false
		s.Speed = 0
	}
	h.mu.Unlock()
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
	h.mu.Lock()
	if s, ok := h.machines[id]; ok {
		s.Setpoint = req.Value
		if s.Running {
			s.Speed = req.Value
		}
	}
	h.mu.Unlock()
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
	h.mu.Lock()
	if s, ok := h.machines[id]; ok {
		s.Mode = req.Mode
	}
	h.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "mode_updated"})
}

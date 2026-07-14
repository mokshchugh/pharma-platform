package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"pharma-platform/internal/store"

	"github.com/go-chi/chi/v5"
)

type ProductionHandler struct {
	store *store.ProductionStore
}

func NewProductionHandler(store *store.ProductionStore) *ProductionHandler {
	return &ProductionHandler{store: store}
}

// GET /api/v1/production - list active productions
func (h *ProductionHandler) List(w http.ResponseWriter, r *http.Request) {
	machineStr := r.URL.Query().Get("machine_id")
	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}
	var machineID int
	if machineStr != "" {
		machineID, _ = strconv.Atoi(machineStr)
	}
	runs := h.store.ListRuns(machineID, limit)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(runs)
}

// GET /api/v1/production/active/{machine_id}
func (h *ProductionHandler) GetActive(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "machine_id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid machine id", http.StatusBadRequest)
		return
	}
	run := h.store.GetActiveRun(id)
	if run == nil {
		http.Error(w, "no active run", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(run)
}

// POST /api/v1/production/start
func (h *ProductionHandler) StartRun(w http.ResponseWriter, r *http.Request) {
	var req struct {
		MachineID   int    `json:"machine_id"`
		BatchID     string `json:"batch_id"`
		ProductName string `json:"product_name"`
		TargetQty   int    `json:"target_qty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	run, err := h.store.CreateRun(req.MachineID, req.BatchID, req.ProductName, req.TargetQty)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(run)
}

// POST /api/v1/production/complete/{id}
func (h *ProductionHandler) CompleteRun(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid run id", http.StatusBadRequest)
		return
	}
	if err := h.store.CompleteRun(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "completed"})
}

// GET /api/v1/downtime?machine_id=X&limit=20
func (h *ProductionHandler) ListDowntime(w http.ResponseWriter, r *http.Request) {
	machineStr := r.URL.Query().Get("machine_id")
	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}
	var machineID int
	if machineStr != "" {
		machineID, _ = strconv.Atoi(machineStr)
	}
	events := h.store.ListDowntime(machineID, limit)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

// POST /api/v1/downtime/start
func (h *ProductionHandler) StartDowntime(w http.ResponseWriter, r *http.Request) {
	var req struct {
		MachineID int    `json:"machine_id"`
		Reason    string `json:"reason"`
		Category  string `json:"category"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	event, err := h.store.StartDowntime(req.MachineID, req.Reason, req.Category)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(event)
}

// POST /api/v1/downtime/end/{id}
func (h *ProductionHandler) EndDowntime(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid event id", http.StatusBadRequest)
		return
	}
	if err := h.store.EndDowntime(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ended"})
}

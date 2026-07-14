package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"pharma-platform/internal/store"

	"github.com/go-chi/chi/v5"
)

type OEEHandler struct {
	store *store.ProductionStore
}

func NewOEEHandler(store *store.ProductionStore) *OEEHandler {
	return &OEEHandler{store: store}
}

// GET /api/v1/oee - all machines OEE
func (h *OEEHandler) List(w http.ResponseWriter, r *http.Request) {
	machines := h.store.GetDashboardSummary()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(machines)
}

// GET /api/v1/oee/{id} - single machine OEE
func (h *OEEHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid machine id", http.StatusBadRequest)
		return
	}
	window := 24 * time.Hour
	if w := r.URL.Query().Get("window"); w != "" {
		if d, err := time.ParseDuration(w); err == nil {
			window = d
		}
	}
	resp := h.store.CalculateOEE(id, window)
	if resp == nil {
		http.Error(w, "machine not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

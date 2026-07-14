package handlers

import (
	"encoding/json"
	"net/http"

	"pharma-platform/internal/models"
	"pharma-platform/internal/store"
)

type DashboardHandler struct {
	productionStore *store.ProductionStore
	alarmStore      *AlarmStore
}

func NewDashboardHandler(productionStore *store.ProductionStore, alarmStore *AlarmStore) *DashboardHandler {
	return &DashboardHandler{productionStore: productionStore, alarmStore: alarmStore}
}

func (h *DashboardHandler) Summary(w http.ResponseWriter, r *http.Request) {
	summary := h.productionStore.GetDashboardSummary()
	if summary == nil {
		summary = &models.DashboardSummary{}
	}
	summary.ActiveAlarms = h.alarmStore.ActiveCount()
	summary.CriticalAlarms = h.alarmStore.CriticalCount()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

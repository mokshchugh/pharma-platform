package handlers

import (
	"encoding/json"
	"net/http"

	"pharma-platform/internal/business"
)

type BusinessAnalyticsHandler struct {
	engine business.Engine
}

func NewBusinessAnalyticsHandler(engine business.Engine) *BusinessAnalyticsHandler {
	return &BusinessAnalyticsHandler{engine: engine}
}

func (h *BusinessAnalyticsHandler) Overview(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.engine.GetOverview())
}

func (h *BusinessAnalyticsHandler) Production(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.engine.GetProductionAnalytics())
}

func (h *BusinessAnalyticsHandler) Quality(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.engine.GetQualityAnalytics())
}

func (h *BusinessAnalyticsHandler) Machines(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.engine.GetMachineAnalytics())
}

func (h *BusinessAnalyticsHandler) Energy(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.engine.GetEnergyAnalytics())
}

func (h *BusinessAnalyticsHandler) Alarms(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.engine.GetAlarmAnalytics())
}

func (h *BusinessAnalyticsHandler) Correlations(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.engine.GetCorrelationAnalysis())
}

func (h *BusinessAnalyticsHandler) Maintenance(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.engine.GetMaintenanceAnalysis())
}

func (h *BusinessAnalyticsHandler) Insights(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.engine.GetInsights())
}

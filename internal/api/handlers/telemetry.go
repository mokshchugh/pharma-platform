package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"pharma-platform/internal/questdb"

	"github.com/go-chi/chi/v5"
)

type TelemetryHandler struct {
	reader *questdb.Reader
}

func NewTelemetryHandler(reader *questdb.Reader) *TelemetryHandler {
	return &TelemetryHandler{reader: reader}
}

func (h *TelemetryHandler) Latest(w http.ResponseWriter, r *http.Request) {
	data, err := h.reader.Latest(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(data)
}

func (h *TelemetryHandler) LatestByPLC(w http.ResponseWriter, r *http.Request) {
	plcID := chi.URLParam(r, "plc_id")

	data, err := h.reader.LatestByPLC(r.Context(), plcID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(data)
}

func (h *TelemetryHandler) LatestByPLCAndTag(w http.ResponseWriter, r *http.Request) {
	plcID := chi.URLParam(r, "plc_id")
	tagID := chi.URLParam(r, "tag_id")

	sample, err := h.reader.LatestByPLCAndTag(r.Context(), plcID, tagID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if sample == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(sample)
}

func (h *TelemetryHandler) History(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	plcID := query.Get("plc_id")
	tagID := query.Get("tag_id")

	start, err := time.Parse(time.RFC3339, query.Get("start"))
	if err != nil {
		http.Error(w, "invalid start", http.StatusBadRequest)
		return
	}

	end, err := time.Parse(time.RFC3339, query.Get("end"))
	if err != nil {
		http.Error(w, "invalid end", http.StatusBadRequest)
		return
	}

	data, err := h.reader.History(r.Context(), plcID, tagID, start, end)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(data)
}

func (h *TelemetryHandler) Aggregate(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	plcID := q.Get("plc_id")
	tagID := q.Get("tag_id")
	interval := q.Get("interval")

	start, err := time.Parse(time.RFC3339, q.Get("start"))
	if err != nil {
		http.Error(w, "invalid start", http.StatusBadRequest)
		return
	}

	end, err := time.Parse(time.RFC3339, q.Get("end"))
	if err != nil {
		http.Error(w, "invalid end", http.StatusBadRequest)
		return
	}

	switch interval {
	case "1m", "1h", "1d", "1w":
	default:
		http.Error(w, "invalid interval (use 1m, 1h, 1d, or 1w)", http.StatusBadRequest)
		return
	}

	data, err := h.reader.Aggregate(r.Context(), plcID, tagID, interval, start, end)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(data)
}

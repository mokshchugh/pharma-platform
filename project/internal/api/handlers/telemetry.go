package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
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
	machineID := chi.URLParam(r, "plc_id")

	data, err := h.reader.LatestByPLC(r.Context(), machineID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(data)
}

func (h *TelemetryHandler) LatestByPLCAndTag(w http.ResponseWriter, r *http.Request) {
	machineID := chi.URLParam(r, "plc_id")
	tagName := chi.URLParam(r, "tag_id")

	sample, err := h.reader.LatestByPLCAndTag(r.Context(), machineID, tagName)
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

	machineID := query.Get("plc_id")
	tagName := query.Get("tag_id")

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

	data, err := h.reader.History(r.Context(), machineID, tagName, start, end)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(data)
}

func (h *TelemetryHandler) Aggregate(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	machineID := q.Get("plc_id")
	tagName := q.Get("tag_id")
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

	data, err := h.reader.Aggregate(r.Context(), machineID, tagName, interval, start, end)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(data)
}

var resolutionAllowlist = map[string]string{
	"raw": "plc_samples",
	"1m":  "plc_samples_1m",
	"1h":  "plc_samples_1h",
	"1d":  "plc_samples_1d",
	"1w":  "plc_samples_1w",
}

func (h *TelemetryHandler) DataStream(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	resolution := q.Get("resolution")
	if resolution == "" {
		http.Error(w, "resolution is required", http.StatusBadRequest)
		return
	}

	table, ok := resolutionAllowlist[resolution]
	if !ok {
		http.Error(w, "invalid resolution (raw, 1m, 1h, 1d, 1w)", http.StatusBadRequest)
		return
	}

	machineID := q.Get("machine")
	tagName := q.Get("plc")

	start, err := time.Parse(time.RFC3339, q.Get("start"))
	if err != nil {
		http.Error(w, "invalid start (use RFC3339)", http.StatusBadRequest)
		return
	}

	end, err := time.Parse(time.RFC3339, q.Get("end"))
	if err != nil {
		http.Error(w, "invalid end (use RFC3339)", http.StatusBadRequest)
		return
	}

	page := 1
	pageSize := 100
	if p := q.Get("page"); p != "" {
		if v, err := parseInt(p); err == nil && v > 0 {
			page = v
		}
	}
	if ps := q.Get("page_size"); ps != "" {
		if v, err := parseInt(ps); err == nil && v > 0 && v <= 1000 {
			pageSize = v
		}
	}

	var resp *questdb.StreamResponse
	if resolution == "raw" {
		resp, err = h.reader.StreamRaw(r.Context(), table, machineID, tagName, start, end, page, pageSize)
	} else {
		resp, err = h.reader.StreamAggregate(r.Context(), table, machineID, tagName, start, end, page, pageSize, resolution)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *TelemetryHandler) DataStreamCSV(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	resolution := q.Get("resolution")
	if resolution == "" {
		http.Error(w, "resolution is required", http.StatusBadRequest)
		return
	}

	table, ok := resolutionAllowlist[resolution]
	if !ok {
		http.Error(w, "invalid resolution (raw, 1m, 1h, 1d, 1w)", http.StatusBadRequest)
		return
	}

	machineID := q.Get("machine")
	tagName := q.Get("plc")

	start, err := time.Parse(time.RFC3339, q.Get("start"))
	if err != nil {
		http.Error(w, "invalid start (use RFC3339)", http.StatusBadRequest)
		return
	}

	end, err := time.Parse(time.RFC3339, q.Get("end"))
	if err != nil {
		http.Error(w, "invalid end (use RFC3339)", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", `attachment; filename="telemetry-export.csv"`)

	if resolution == "raw" {
		rows, err := h.reader.StreamRawAll(r.Context(), table, machineID, tagName, start, end)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintln(w, "timestamp,machine_id,machine_name,tag_name,value,quality")
		for _, row := range rows {
			fmt.Fprintf(w, "%s,%s,%s,%s,%.4f,%d\n",
				escapeCSV(row.Timestamp),
				escapeCSV(row.MachineID),
				escapeCSV(row.MachineName),
				escapeCSV(row.TagName),
				row.Value,
				row.Quality,
			)
		}
	} else {
		rows, err := h.reader.StreamAggregateAll(r.Context(), table, machineID, tagName, start, end)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintln(w, "timestamp,machine_id,machine_name,tag_name,min_value,max_value,avg_value,sample_count")
		for _, row := range rows {
			fmt.Fprintf(w, "%s,%s,%s,%s,%.4f,%.4f,%.4f,%d\n",
				escapeCSV(row.Timestamp),
				escapeCSV(row.MachineID),
				escapeCSV(row.MachineName),
				escapeCSV(row.TagName),
				row.MinValue,
				row.MaxValue,
				row.AvgValue,
				row.SampleCount,
			)
		}
	}
}

func escapeCSV(s string) string {
	needsQuoting := false
	for _, c := range s {
		if c == ',' || c == '"' || c == '\n' || c == '\r' {
			needsQuoting = true
			break
		}
	}
	if !needsQuoting {
		return s
	}
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}

func parseInt(s string) (int, error) {
	var v int
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("not a number")
		}
		v = v*10 + int(c-'0')
	}
	return v, nil
}

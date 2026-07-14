package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"pharma-platform/internal/models"
	"pharma-platform/internal/questdb"
	"pharma-platform/internal/store"

	"github.com/go-chi/chi/v5"
)

type TelemetryTagResponse struct {
	TagID    int    `json:"tag_id"`
	TagName  string `json:"tag_name"`
	DataType string `json:"data_type"`
	Unit     string `json:"unit"`
}

type AnalyticsPoint struct {
	Timestamp   string  `json:"timestamp"`
	AvgValue    float64 `json:"avg_value"`
	MinValue    float64 `json:"min_value"`
	MaxValue    float64 `json:"max_value"`
	SampleCount int64   `json:"sample_count"`
}

type TagAnalyticsResponse struct {
	TagID            int               `json:"tag_id"`
	TagName          string            `json:"tag_name"`
	DataType         string            `json:"data_type"`
	Unit             string            `json:"unit"`
	Current          *AnalyticsPoint   `json:"current,omitempty"`
	LatestValue      *float64          `json:"latest_value,omitempty"`
	Series           []AnalyticsPoint  `json:"series,omitempty"`
	TotalSampleCount int64            `json:"total_sample_count"`
	WindowAvg        float64           `json:"window_avg"`
	WindowMin        float64           `json:"window_min"`
	WindowMax        float64           `json:"window_max"`
}

type AnalyticsResponse struct {
	Tags []TagAnalyticsResponse `json:"tags"`
}

type AnalyticsHandler struct {
	tagStore *store.TagStore
	reader   *questdb.Reader
}

func NewAnalyticsHandler(tagStore *store.TagStore, reader *questdb.Reader) *AnalyticsHandler {
	return &AnalyticsHandler{tagStore: tagStore, reader: reader}
}

func (h *AnalyticsHandler) GetTelemetry(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	machineID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid machine id", http.StatusBadRequest)
		return
	}

	tags := h.tagStore.GetTagsByMachineID(machineID)
	resp := make([]TelemetryTagResponse, 0, len(tags))
	for _, t := range tags {
		resp = append(resp, TelemetryTagResponse{
			TagID:    parseTagNumericID(t.ID),
			TagName:  t.Name,
			DataType: t.DataType.String(),
			Unit:     t.Unit,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *AnalyticsHandler) GetAnalytics(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	machineID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid machine id", http.StatusBadRequest)
		return
	}

	resolution := r.URL.Query().Get("resolution")
	switch resolution {
	case "1m", "1h", "1d", "1w":
	default:
		http.Error(w, "invalid resolution (use 1m, 1h, 1d, or 1w)", http.StatusBadRequest)
		return
	}

	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	from, err := time.Parse(time.RFC3339, fromStr)
	if err != nil {
		http.Error(w, "invalid from (use RFC3339)", http.StatusBadRequest)
		return
	}

	to, err := time.Parse(time.RFC3339, toStr)
	if err != nil {
		http.Error(w, "invalid to (use RFC3339)", http.StatusBadRequest)
		return
	}

	tags := h.tagStore.GetTagsByMachineID(machineID)
	qdbID := strconv.Itoa(machineID)

	resp := AnalyticsResponse{
		Tags: make([]TagAnalyticsResponse, 0, len(tags)),
	}

	for _, t := range tags {
		tagResp := TagAnalyticsResponse{
			TagID:    parseTagNumericID(t.ID),
			TagName:  t.Name,
			DataType: t.DataType.String(),
			Unit:     t.Unit,
		}

		if isAnalogTag(t.DataType) {
			latest, err := h.reader.LatestFromView(r.Context(), qdbID, resolution, t.Name)
			if err == nil && latest != nil {
				tagResp.Current = aggregateRowToPoint(latest)
			}

			series, err := h.reader.SeriesFromView(r.Context(), qdbID, resolution, t.Name, from, to)
			if err == nil && len(series) > 0 {
				var total int64
				var weightedSum float64
				windowMin := series[0].MinValue
				windowMax := series[0].MaxValue

				tagResp.Series = make([]AnalyticsPoint, 0, len(series))
				for _, s := range series {
					total += s.SampleCount
					weightedSum += s.AvgValue * float64(s.SampleCount)
					if s.MinValue < windowMin {
						windowMin = s.MinValue
					}
					if s.MaxValue > windowMax {
						windowMax = s.MaxValue
					}
					tagResp.Series = append(tagResp.Series, *aggregateRowToPoint(&s))
				}

				tagResp.TotalSampleCount = total
				if total > 0 {
					tagResp.WindowAvg = weightedSum / float64(total)
				}
				tagResp.WindowMin = windowMin
				tagResp.WindowMax = windowMax
			}
		} else if t.DataType == models.DataTypeBool {
			sample, err := h.reader.LatestByPLCAndTag(r.Context(), qdbID, t.Name)
			if err == nil && sample != nil {
				val := toFloat64(sample.Value)
				tagResp.LatestValue = &val
			}
		}

		resp.Tags = append(resp.Tags, tagResp)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func isAnalogTag(dt models.DataType) bool {
	switch dt {
	case models.DataTypeBool, models.DataTypeString, models.DataTypeBytes:
		return false
	default:
		return true
	}
}

func toFloat64(v any) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case int32:
		return float64(val)
	default:
		return 0
	}
}

func aggregateRowToPoint(row *questdb.AggregateRow) *AnalyticsPoint {
	if row == nil {
		return nil
	}
	return &AnalyticsPoint{
		Timestamp:   row.Timestamp,
		AvgValue:    row.AvgValue,
		MinValue:    row.MinValue,
		MaxValue:    row.MaxValue,
		SampleCount: row.SampleCount,
	}
}

func parseTagNumericID(id string) int {
	parts := strings.SplitN(id, "-", 2)
	if len(parts) == 2 {
		if n, err := strconv.Atoi(parts[1]); err == nil {
			return n
		}
	}
	if n, err := strconv.Atoi(id); err == nil {
		return n
	}
	return 0
}



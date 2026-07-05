package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"pharma-platform/internal/questdb"
)

type TelemetryHandler struct {
	reader *questdb.Reader
}

func NewTelemetryHandler(
	reader *questdb.Reader,
) *TelemetryHandler {
	return &TelemetryHandler{
		reader: reader,
	}
}

func (h *TelemetryHandler) Latest(
	w http.ResponseWriter,
	r *http.Request,
) {
	data, err := h.reader.Latest(r.Context())
	if err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	_ = json.NewEncoder(w).Encode(data)
}

func (h *TelemetryHandler) History(
	w http.ResponseWriter,
	r *http.Request,
) {
	query := r.URL.Query()

	plcID := query.Get("plc_id")
	tagID := query.Get("tag_id")

	start, err := time.Parse(
		time.RFC3339,
		query.Get("start"),
	)
	if err != nil {
		http.Error(w, "invalid start", http.StatusBadRequest)
		return
	}

	end, err := time.Parse(
		time.RFC3339,
		query.Get("end"),
	)
	if err != nil {
		http.Error(w, "invalid end", http.StatusBadRequest)
		return
	}

	data, err := h.reader.History(
		r.Context(),
		plcID,
		tagID,
		start,
		end,
	)
	if err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	_ = json.NewEncoder(w).Encode(data)
}

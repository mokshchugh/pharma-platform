package handlers

import (
	"encoding/json"
	"net/http"

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

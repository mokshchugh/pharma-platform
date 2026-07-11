package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"pharma-platform/internal/models"

	"github.com/go-chi/chi/v5"
)

type PLCStore interface {
	GetPLCs() []models.PLC
	GetPLC(id string) *models.PLC
	GetTagsByPLC(plcID string) []models.Tag
}

type PLCResponse struct {
	ID        string          `json:"id"`
	MachineID int             `json:"machine_id"`
	Name      string          `json:"machine_name"`
	Driver    models.DriverType `json:"driver"`
}

type PLCStatusResponse struct {
	ID         string `json:"id"`
	Connected  bool   `json:"connected"`
	TagsCount  int    `json:"tags_count"`
	TagsActive int    `json:"tags_active"`
}

type PLCHandler struct {
	store PLCStore
}

func NewPLCHandler(store PLCStore) *PLCHandler {
	return &PLCHandler{store: store}
}

func (h *PLCHandler) List(w http.ResponseWriter, r *http.Request) {
	plcs := h.store.GetPLCs()
	resp := make([]PLCResponse, 0, len(plcs))

	for _, p := range plcs {
		resp = append(resp, PLCResponse{
			ID:        p.ID,
			MachineID: parseMachineNumericID(p.ID),
			Name:      p.Name,
			Driver:    p.Driver,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *PLCHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "plc_id")
	plc := h.store.GetPLC(id)
	if plc == nil {
		http.Error(w, "plc not found", http.StatusNotFound)
		return
	}

	resp := PLCResponse{
		ID:        plc.ID,
		MachineID: parseMachineNumericID(plc.ID),
		Name:      plc.Name,
		Driver:    plc.Driver,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *PLCHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "plc_id")
	plc := h.store.GetPLC(id)
	if plc == nil {
		http.Error(w, "plc not found", http.StatusNotFound)
		return
	}

	tags := h.store.GetTagsByPLC(id)
	active := 0
	for _, t := range tags {
		if t.Enabled {
			active++
		}
	}

	resp := PLCStatusResponse{
		ID:         plc.ID,
		Connected:  plc.Enabled,
		TagsCount:  len(tags),
		TagsActive: active,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func parseMachineNumericID(id string) int {
	parts := strings.SplitN(id, "-", 2)
	if len(parts) != 2 {
		return 0
	}
	n, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0
	}
	return n
}

func (h *PLCHandler) ListTags(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "plc_id")
	plc := h.store.GetPLC(id)
	if plc == nil {
		http.Error(w, "plc not found", http.StatusNotFound)
		return
	}

	tags := h.store.GetTagsByPLC(id)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(tags)
}

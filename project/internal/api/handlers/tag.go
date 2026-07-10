package handlers

import (
	"encoding/json"
	"net/http"

	"pharma-platform/internal/models"

	"github.com/go-chi/chi/v5"
)

type TagStore interface {
	GetTags() []models.Tag
	GetTag(id string) *models.Tag
}

type TagHandler struct {
	store TagStore
}

type TagResponse struct {
	ID           string `json:"id"`
	PLCID        string `json:"plc_id"`
	Name         string `json:"name"`
	MachineID    int    `json:"machine_id"`
	MachineName  string `json:"machine_name"`
	Address      string `json:"address"`
	DataType     string `json:"data_type"`
	Unit         string `json:"unit"`
	PollInterval string `json:"poll_interval"`
	Enabled      bool   `json:"enabled"`
}

func NewTagHandler(store TagStore) *TagHandler {
	return &TagHandler{store: store}
}

func tagToResponse(t models.Tag) TagResponse {
	return TagResponse{
		ID:           t.ID,
		PLCID:        t.PLCID,
		Name:         t.Name,
		MachineID:    t.MachineID,
		MachineName:  t.MachineName,
		Address:      t.Address,
		DataType:     t.DataType.String(),
		Unit:         t.Unit,
		PollInterval: t.PollInterval.String(),
		Enabled:      t.Enabled,
	}
}

func (h *TagHandler) List(w http.ResponseWriter, r *http.Request) {
	tags := h.store.GetTags()
	resp := make([]TagResponse, 0, len(tags))

	for _, t := range tags {
		resp = append(resp, tagToResponse(t))
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *TagHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "tag_id")
	tag := h.store.GetTag(id)
	if tag == nil {
		http.Error(w, "tag not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(tagToResponse(*tag))
}

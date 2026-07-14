package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"pharma-platform/internal/questdb"
	"pharma-platform/internal/store"

	"github.com/go-chi/chi/v5"
)

type MachineStore interface {
	GetAllMachines() ([]store.MachineRow, error)
	GetMachine(id int) (*store.MachineRow, error)
}

type MachineListResponse struct {
	ID               int     `json:"id"`
	MachineName      string  `json:"machine_name"`
	PLCMake          string  `json:"plc_make"`
	PLCModel         string  `json:"plc_model"`
	Status           string  `json:"status"`
	CollectionStatus string  `json:"collection_status"`
	LastSample       *string `json:"last_sample"`
	ConfiguredTags   int     `json:"configured_tags"`
	EnabledTags      int     `json:"enabled_tags"`
}

type MachineHandler struct {
	store  MachineStore
	reader *questdb.Reader
}

func NewMachineHandler(store MachineStore, reader *questdb.Reader) *MachineHandler {
	return &MachineHandler{store: store, reader: reader}
}

func (h *MachineHandler) List(w http.ResponseWriter, r *http.Request) {
	machines, err := h.store.GetAllMachines()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tsMap := make(map[int]time.Time)
	if h.reader != nil {
		timestamps, err := h.reader.LatestTimestamps(r.Context())
		if err == nil {
			for _, mt := range timestamps {
				id := parseNumericID(mt.MachineID)
				if id > 0 {
					tsMap[id] = mt.Timestamp
				}
			}
		}
	}

	resp := make([]MachineListResponse, 0, len(machines))
	for _, m := range machines {
		item := MachineListResponse{
			ID:               m.ID,
			MachineName:      m.MachineName,
			PLCMake:          m.Brand,
			PLCModel:         m.Model,
			Status:           "UNKNOWN",
			CollectionStatus: "COLLECTING",
			ConfiguredTags:   m.ConfiguredTags,
			EnabledTags:      m.EnabledTags,
		}
		if ts, ok := tsMap[m.ID]; ok {
			s := ts.UTC().Format(time.RFC3339)
			item.LastSample = &s
		}
		resp = append(resp, item)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *MachineHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid machine id", http.StatusBadRequest)
		return
	}

	machine, err := h.store.GetMachine(id)
	if err != nil {
		http.Error(w, "machine not found", http.StatusNotFound)
		return
	}

	resp := MachineListResponse{
		ID:               machine.ID,
		MachineName:      machine.MachineName,
		PLCMake:          machine.Brand,
		PLCModel:         machine.Model,
		Status:           "UNKNOWN",
		CollectionStatus: "COLLECTING",
		ConfiguredTags:   machine.ConfiguredTags,
		EnabledTags:      machine.EnabledTags,
	}

	if h.reader != nil {
		timestamps, err := h.reader.LatestTimestamps(r.Context())
		if err == nil {
			for _, mt := range timestamps {
				if parseNumericID(mt.MachineID) == id {
					s := mt.Timestamp.UTC().Format(time.RFC3339)
					resp.LastSample = &s
					break
				}
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func parseNumericID(s string) int {
	if n, err := strconv.Atoi(s); err == nil {
		return n
	}
	parts := strings.SplitN(s, "-", 2)
	if len(parts) != 2 {
		return 0
	}
	n, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0
	}
	return n
}

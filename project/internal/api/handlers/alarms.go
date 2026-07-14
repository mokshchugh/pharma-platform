package handlers

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
)

type Alarm struct {
	ID        string     `json:"id"`
	PLCID     string     `json:"plc_id"`
	TagID     string     `json:"tag_id,omitempty"`
	Message   string     `json:"message"`
	Severity  string     `json:"severity"`
	Active    bool       `json:"active"`
	CreatedAt time.Time  `json:"created_at"`
	AckAt     *time.Time `json:"acknowledged_at,omitempty"`
}

type AlarmStore struct {
	mu      sync.RWMutex
	alarms  []Alarm
	counter int
}

func NewAlarmStore() *AlarmStore {
	return &AlarmStore{}
}

func (s *AlarmStore) List() []Alarm {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]Alarm, len(s.alarms))
	copy(result, s.alarms)
	return result
}

func (s *AlarmStore) ListActive() []Alarm {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []Alarm
	for _, a := range s.alarms {
		if a.Active {
			result = append(result, a)
		}
	}
	return result
}

func (s *AlarmStore) ActiveCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count := 0
	for _, a := range s.alarms {
		if a.Active {
			count++
		}
	}
	return count
}

func (s *AlarmStore) Acknowledge(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	for i := range s.alarms {
		if s.alarms[i].ID == id {
			s.alarms[i].Active = false
			s.alarms[i].AckAt = &now
			return
		}
	}
}

func (s *AlarmStore) CriticalCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count := 0
	for _, a := range s.alarms {
		if a.Active && a.Severity == "critical" {
			count++
		}
	}
	return count
}

type AlarmHandler struct {
	store *AlarmStore
}

func NewAlarmHandler(store *AlarmStore) *AlarmHandler {
	return &AlarmHandler{store: store}
}

func (h *AlarmHandler) List(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(h.store.List())
}

func (h *AlarmHandler) Acknowledge(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	h.store.Acknowledge(idStr)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "acknowledged"})
}

func (h *AlarmHandler) ListActive(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(h.store.ListActive())
}

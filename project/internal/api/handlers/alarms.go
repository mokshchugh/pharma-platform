package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"pharma-platform/internal/questdb"
	"pharma-platform/internal/store"

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
	reader *questdb.Reader
	acks   *store.AlarmAckStore
}

func NewAlarmStore(reader *questdb.Reader, acks *store.AlarmAckStore) *AlarmStore {
	return &AlarmStore{reader: reader, acks: acks}
}

func alarmKey(machineID, tagName string, ts time.Time) string {
	return machineID + "|" + tagName + "|" + ts.UTC().Format(time.RFC3339Nano)
}

func (s *AlarmStore) all() []Alarm {
	rows, err := s.reader.ListAlarms(context.Background(), 500)
	if err != nil {
		return nil
	}

	ackedSet := s.acks.AckedSet()

	alarms := make([]Alarm, 0, len(rows))
	for _, row := range rows {
		key := alarmKey(row.MachineID, row.TagName, row.Timestamp)

		alarm := Alarm{
			ID:        key,
			PLCID:     row.MachineID,
			TagID:     row.TagName,
			Message:   row.Message,
			Severity:  row.Severity,
			Active:    true,
			CreatedAt: row.Timestamp,
		}

		if ackedAt, ok := ackedSet[key]; ok {
			alarm.Active = false
			at := ackedAt
			alarm.AckAt = &at
		}

		alarms = append(alarms, alarm)
	}

	return alarms
}

func (s *AlarmStore) List() []Alarm {
	return s.all()
}

func (s *AlarmStore) ListActive() []Alarm {
	var active []Alarm
	for _, a := range s.all() {
		if a.Active {
			active = append(active, a)
		}
	}
	return active
}

func (s *AlarmStore) ActiveCount() int {
	return len(s.ListActive())
}

func (s *AlarmStore) CriticalCount() int {
	count := 0
	for _, a := range s.ListActive() {
		if a.Severity == "critical" {
			count++
		}
	}
	return count
}

func (s *AlarmStore) Acknowledge(id string) {
	parts := strings.SplitN(id, "|", 3)
	if len(parts) != 3 {
		return
	}

	ts, err := time.Parse(time.RFC3339Nano, parts[2])
	if err != nil {
		return
	}

	s.acks.Ack(parts[0], parts[1], ts)
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

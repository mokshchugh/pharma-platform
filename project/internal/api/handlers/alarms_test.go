package handlers

import (
	"testing"

	"pharma-platform/internal/postgres"
	"pharma-platform/internal/store"
)

func newTestAlarmStore() *AlarmStore {
	// An unconnected client: store methods check DB() == nil and return
	// cleanly rather than touching the network, so this exercises the
	// id-parsing logic without needing a live Postgres instance.
	return &AlarmStore{acks: store.NewAlarmAckStore(postgres.New(postgres.Config{}))}
}

func TestAcknowledgeParsesWellFormedID(t *testing.T) {
	s := newTestAlarmStore()

	// A real alarm ID, as produced by alarmKey: machineID|tagName|RFC3339Nano
	// timestamp. This is the decoded form the handler must pass in — the
	// raw HTTP path segment is percent-encoded and must be unescaped first
	// (see AlarmHandler.Acknowledge).
	err := s.Acknowledge("1|Alarm_Status|2026-08-24T00:01:38.638737Z")
	if err != nil {
		t.Fatalf("expected a well-formed id to parse without error, got: %v", err)
	}
}

func TestAcknowledgeRejectsMalformedID(t *testing.T) {
	s := newTestAlarmStore()

	for _, id := range []string{"", "onlyonepart", "two|parts"} {
		if err := s.Acknowledge(id); err == nil {
			t.Errorf("expected error for malformed id %q, got nil", id)
		}
	}
}

func TestAcknowledgeRejectsMalformedTimestamp(t *testing.T) {
	s := newTestAlarmStore()

	if err := s.Acknowledge("1|Alarm_Status|not-a-timestamp"); err == nil {
		t.Error("expected error for malformed timestamp, got nil")
	}
}

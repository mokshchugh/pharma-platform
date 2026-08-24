package store

import (
	"fmt"
	"strconv"
	"time"

	"pharma-platform/internal/postgres"
)

type AlarmAckStore struct {
	client *postgres.Client
}

func NewAlarmAckStore(client *postgres.Client) *AlarmAckStore {
	return &AlarmAckStore{client: client}
}

// Ack records an acknowledgement. machineID is the string form used
// throughout the telemetry pipeline (QuestDB has no integer FKs), but is
// stored as an INTEGER referencing machines(id) here since this table
// lives in Postgres alongside every other machine-scoped table.
func (s *AlarmAckStore) Ack(machineID, tagName string, alarmTS time.Time) error {
	db := s.client.DB()
	if db == nil {
		return nil
	}

	id, err := strconv.Atoi(machineID)
	if err != nil {
		return fmt.Errorf("machine id %q is not numeric: %w", machineID, err)
	}

	_, err = db.Exec(
		`INSERT INTO alarm_acks (machine_id, tag_name, alarm_ts) VALUES ($1, $2, $3)`,
		id, tagName, alarmTS,
	)
	return err
}

// AckedSet returns a set of acknowledged alarms keyed by
// "<machine_id>|<tag_name>|<alarm_ts RFC3339Nano UTC>" mapped to the ack timestamp.
func (s *AlarmAckStore) AckedSet() map[string]time.Time {
	db := s.client.DB()
	if db == nil {
		return nil
	}

	rows, err := db.Query(`SELECT machine_id, tag_name, alarm_ts, acked_at FROM alarm_acks`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	acked := make(map[string]time.Time)
	for rows.Next() {
		var machineID int
		var tagName string
		var alarmTS, ackedAt time.Time
		if err := rows.Scan(&machineID, &tagName, &alarmTS, &ackedAt); err != nil {
			continue
		}
		key := strconv.Itoa(machineID) + "|" + tagName + "|" + alarmTS.UTC().Format(time.RFC3339Nano)
		acked[key] = ackedAt
	}

	return acked
}

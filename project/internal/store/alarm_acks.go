package store

import (
	"time"

	"pharma-platform/internal/postgres"
)

type AlarmAckStore struct {
	client *postgres.Client
}

func NewAlarmAckStore(client *postgres.Client) *AlarmAckStore {
	return &AlarmAckStore{client: client}
}

func (s *AlarmAckStore) Ack(machineID, tagName string, alarmTS time.Time) error {
	db := s.client.DB()
	if db == nil {
		return nil
	}

	_, err := db.Exec(
		`INSERT INTO alarm_acks (machine_id, tag_name, alarm_ts) VALUES ($1, $2, $3)`,
		machineID, tagName, alarmTS,
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
		var machineID, tagName string
		var alarmTS, ackedAt time.Time
		if err := rows.Scan(&machineID, &tagName, &alarmTS, &ackedAt); err != nil {
			continue
		}
		key := machineID + "|" + tagName + "|" + alarmTS.UTC().Format(time.RFC3339Nano)
		acked[key] = ackedAt
	}

	return acked
}

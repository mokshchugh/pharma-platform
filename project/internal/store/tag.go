package store

import (
	"fmt"
	"time"

	"pharma-platform/internal/models"
	"pharma-platform/internal/postgres"
)

type TagStore struct {
	client *postgres.Client
}

func NewTagStore(client *postgres.Client) *TagStore {
	return &TagStore{client: client}
}

func (s *TagStore) GetTags() []models.Tag {
	db := s.client.DB()
	if db == nil {
		return nil
	}

	rows, err := db.Query(
		`SELECT t.id, t.tag_name, t.data_type, t.address, COALESCE(t.unit, ''), COALESCE(t.scale_factor, 1.0), t.enabled, t.machine_id, m.machine_name
		 FROM tags t
		 JOIN machines m ON m.id = t.machine_id
		 WHERE t.enabled = true
		 ORDER BY t.id`,
	)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var tags []models.Tag
	for rows.Next() {
		var t models.Tag
		var tagID, machineID int
		var dtype string
		if err := rows.Scan(&tagID, &t.Name, &dtype, &t.Address, &t.Unit, &t.ScaleFactor, &t.Enabled, &machineID, &t.MachineName); err != nil {
			continue
		}
		t.ID = fmt.Sprintf("tag-%d", tagID)
		t.PLCID = fmt.Sprintf("machine-%d", machineID)
		t.MachineID = machineID
		t.DataType = parseDataType(dtype)
		t.PollInterval = 100 * time.Millisecond
		tags = append(tags, t)
	}

	return tags
}

func (s *TagStore) GetTag(id string) *models.Tag {
	db := s.client.DB()
	if db == nil {
		return nil
	}

	dbID := machineIDFromString(id)

	var t models.Tag
	var tagID, machineID int
	var dtype string

	err := db.QueryRow(
		`SELECT t.id, t.tag_name, t.data_type, t.address, COALESCE(t.unit, ''), COALESCE(t.scale_factor, 1.0), t.enabled, t.machine_id, m.machine_name
		 FROM tags t
		 JOIN machines m ON m.id = t.machine_id
		 WHERE t.id = $1`,
		dbID,
	).Scan(&tagID, &t.Name, &dtype, &t.Address, &t.Unit, &t.ScaleFactor, &t.Enabled, &machineID, &t.MachineName)
	if err != nil {
		return nil
	}

	t.ID = fmt.Sprintf("tag-%d", tagID)
	t.PLCID = fmt.Sprintf("machine-%d", machineID)
	t.MachineID = machineID
	t.DataType = parseDataType(dtype)
	t.PollInterval = 100 * time.Millisecond
	return &t
}

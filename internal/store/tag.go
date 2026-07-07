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
		`SELECT t.id, t.tag_name, t.data_type, t.address, COALESCE(t.unit, ''), COALESCE(t.scale_factor, 1.0), t.enabled, t.machine_id
		 FROM tags t
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
		if err := rows.Scan(&tagID, &t.Name, &dtype, &t.Address, &t.Unit, &t.ScaleFactor, &t.Enabled, &machineID); err != nil {
			continue
		}
		t.ID = fmt.Sprintf("tag-%d", tagID)
		t.PLCID = fmt.Sprintf("machine-%d", machineID)
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
		`SELECT t.id, t.tag_name, t.data_type, t.address, COALESCE(t.unit, ''), COALESCE(t.scale_factor, 1.0), t.enabled, t.machine_id
		 FROM tags t
		 WHERE t.id = $1`,
		dbID,
	).Scan(&tagID, &t.Name, &dtype, &t.Address, &t.Unit, &t.ScaleFactor, &t.Enabled, &machineID)
	if err != nil {
		return nil
	}

	t.ID = fmt.Sprintf("tag-%d", tagID)
	t.PLCID = fmt.Sprintf("machine-%d", machineID)
	t.DataType = parseDataType(dtype)
	t.PollInterval = 100 * time.Millisecond
	return &t
}

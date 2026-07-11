package store

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"pharma-platform/internal/models"
	"pharma-platform/internal/postgres"
)

type MachineStore struct {
	client *postgres.Client
}

func NewMachineStore(client *postgres.Client) *MachineStore {
	return &MachineStore{client: client}
}

func (s *MachineStore) GetPLCs() []models.PLC {
	db := s.client.DB()
	if db == nil {
		return nil
	}

	rows, err := db.Query(
		`SELECT id, machine_name, protocol, COALESCE(ip_address, ''), COALESCE(port, 0), enabled
		 FROM machines WHERE enabled = true ORDER BY id`,
	)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var plcs []models.PLC
	for rows.Next() {
		var p models.PLC
		var dbID int
		var driverStr string
		if err := rows.Scan(&dbID, &p.Name, &driverStr, &p.IPAddress, &p.Port, &p.Enabled); err != nil {
			continue
		}
		p.ID = fmt.Sprintf("machine-%d", dbID)
		p.Driver = models.DriverType(driverStr)
		plcs = append(plcs, p)
	}

	return plcs
}

func (s *MachineStore) GetPLC(id string) *models.PLC {
	db := s.client.DB()
	if db == nil {
		return nil
	}

	var dbID int
	var p models.PLC
	var driverStr string

	err := db.QueryRow(
		`SELECT id, machine_name, protocol, COALESCE(ip_address, ''), COALESCE(port, 0), enabled
		 FROM machines WHERE id = $1`,
		machineIDFromString(id),
	).Scan(&dbID, &p.Name, &driverStr, &p.IPAddress, &p.Port, &p.Enabled)
	if err != nil {
		return nil
	}

	p.ID = fmt.Sprintf("machine-%d", dbID)
	p.Driver = models.DriverType(driverStr)
	return &p
}

func (s *MachineStore) GetTagsByPLC(plcID string) []models.Tag {
	db := s.client.DB()
	if db == nil {
		return nil
	}

	dbID := machineIDFromString(plcID)

	rows, err := db.Query(
		`SELECT t.id, t.tag_name, t.data_type, t.address, COALESCE(t.unit, ''), COALESCE(t.scale_factor, 1.0), t.enabled, m.machine_name
		 FROM tags t
		 JOIN machines m ON m.id = t.machine_id
		 WHERE t.machine_id = $1 AND t.enabled = true ORDER BY t.id`,
		dbID,
	)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var tags []models.Tag
	for rows.Next() {
		var t models.Tag
		var tagID int
		var dtype string
		if err := rows.Scan(&tagID, &t.Name, &dtype, &t.Address, &t.Unit, &t.ScaleFactor, &t.Enabled, &t.MachineName); err != nil {
			continue
		}
		t.ID = fmt.Sprintf("tag-%d", tagID)
		t.PLCID = plcID
		t.MachineID = dbID
		t.DataType = parseDataType(dtype)
		t.PollInterval = 100 * time.Millisecond
		tags = append(tags, t)
	}

	return tags
}

func machineIDFromString(s string) int {
	parts := strings.SplitN(s, "-", 2)
	if len(parts) != 2 {
		return 0
	}
	id, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0
	}
	return id
}

func parseDataType(s string) models.DataType {
	switch strings.ToLower(s) {
	case "bool":
		return models.DataTypeBool
	case "int16":
		return models.DataTypeInt16
	case "int32", "dint":
		return models.DataTypeInt32
	case "int64":
		return models.DataTypeInt64
	case "uint16":
		return models.DataTypeUint16
	case "uint32":
		return models.DataTypeUint32
	case "uint64":
		return models.DataTypeUint64
	case "float32", "real":
		return models.DataTypeFloat32
	case "float64":
		return models.DataTypeFloat64
	case "string":
		return models.DataTypeString
	default:
		return models.DataTypeFloat64
	}
}

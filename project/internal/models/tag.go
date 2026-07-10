package models

import "time"

// Tag represents a PLC variable that the collector should read.
type Tag struct {
	ID           string        `yaml:"id" json:"id"`
	PLCID        string        `yaml:"plc_id" json:"plc_id"`
	Name         string        `yaml:"name" json:"name"`
	MachineID    int           `yaml:"machine_id" json:"machine_id"`
	MachineName  string        `yaml:"machine_name" json:"machine_name"`
	Address      string        `yaml:"address" json:"address"`
	DataType     DataType      `yaml:"data_type" json:"data_type"`
	Unit         string        `yaml:"unit" json:"unit"`
	ScaleFactor  float64       `yaml:"scale_factor" json:"scale_factor"`
	PollInterval time.Duration `yaml:"poll_interval" json:"poll_interval"`
	Enabled      bool          `yaml:"enabled" json:"enabled"`
}

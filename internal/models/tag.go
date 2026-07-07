package models

import "time"

// Tag represents a PLC variable that the collector should read.
type Tag struct {
	ID           string        `yaml:"id"`
	PLCID        string        `yaml:"plc_id"`
	Name         string        `yaml:"name"`
	Address      string        `yaml:"address"`
	DataType     DataType      `yaml:"data_type"`
	Unit         string        `yaml:"unit"`
	ScaleFactor  float64       `yaml:"scale_factor"`
	PollInterval time.Duration `yaml:"poll_interval"`
	Enabled      bool          `yaml:"enabled"`
}

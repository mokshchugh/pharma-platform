package models

import "time"

// Tag represents a PLC variable that the collector should read.
type Tag struct {
	ID           string
	PLCID        string
	Name         string
	Address      string
	DataType     DataType
	PollInterval time.Duration
	Enabled      bool
}

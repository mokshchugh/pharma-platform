package models

import "time"

// Sample represents a single telemetry value collected from a PLC.
type Sample struct {
	Timestamp time.Time
	PLCID     string
	TagID     string
	Value     any
	Quality   Quality
}

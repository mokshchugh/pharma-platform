package models

import "time"

// Sample represents a single telemetry value collected from a PLC.
type Sample struct {
	Timestamp time.Time

	MachineID   string
	MachineName string
	TagName     string

	Value any

	Quality Quality
}

package models

// Quality represents the validity of a telemetry sample.
type Quality uint8

const (
	QualityGood Quality = iota
	QualityBad
	QualityUncertain
	QualityDisconnected
	QualityTimeout
)

package models

// PLC represents a programmable logic controller configured in the plant.
type PLC struct {
	ID        string
	Name      string
	Driver    string
	IPAddress string
	Port      uint16
	Enabled   bool
}

package models

// PLC represents a programmable logic controller configured in the plant.
type PLC struct {
	ID        string `yaml:"id"`
	Name      string `yaml:"name"`
	Driver    string `yaml:"driver"`
	IPAddress string `yaml:"ip_address"`
	Port      uint16 `yaml:"port"`
	Enabled   bool   `yaml:"enabled"`
}

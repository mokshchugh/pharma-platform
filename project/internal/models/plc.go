package models

// PLC represents a programmable logic controller configured in the plant.
type PLC struct {
	ID        string     `yaml:"id" json:"id"`
	Name      string     `yaml:"name" json:"machine_name"`
	Driver    DriverType `yaml:"driver" json:"driver"`
	IPAddress string     `yaml:"ip_address" json:"ip_address"`
	Port      uint16     `yaml:"port" json:"port"`
	Enabled   bool       `yaml:"enabled" json:"enabled"`
}

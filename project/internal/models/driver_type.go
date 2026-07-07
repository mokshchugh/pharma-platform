package models

type DriverType string

const (
	DriverOPCUA        DriverType = "opcua"
	DriverModbus       DriverType = "modbus"
	DriverS7           DriverType = "s7"
	DriverMitsubishiMC DriverType = "mc"
	DriverFINS         DriverType = "fins"
	DriverEtherNetIP   DriverType = "ethernetip"
)

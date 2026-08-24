// Package registry constructs a plc.Driver for a configured PLC based on
// its DriverType. It's a separate package from internal/plc so that driver
// implementations (which import internal/plc to satisfy the Driver
// interface) can be imported here without an import cycle.
package registry

import (
	"fmt"

	"pharma-platform/internal/models"
	"pharma-platform/internal/plc"
	"pharma-platform/internal/plc/drivers/opcua"
)

// New constructs a plc.Driver for the given PLC configuration.
//
// Only DriverOPCUA is implemented today. Every other DriverType is a
// recognized, intentional placeholder — it returns a clear error rather
// than silently doing nothing, so callers know the protocol isn't
// supported yet rather than getting an unexplained no-op connection.
func New(p models.PLC) (plc.Driver, error) {
	switch p.Driver {
	case models.DriverOPCUA:
		return opcua.New(opcua.NewConfig(p)), nil

	case models.DriverModbus, models.DriverS7, models.DriverMitsubishiMC,
		models.DriverFINS, models.DriverEtherNetIP:
		return nil, fmt.Errorf("driver %q is not implemented yet (only %q is currently supported)", p.Driver, models.DriverOPCUA)

	default:
		return nil, fmt.Errorf("unknown driver type %q", p.Driver)
	}
}

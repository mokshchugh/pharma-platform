package registry

import (
	"testing"

	"pharma-platform/internal/models"
)

func TestNewOPCUA(t *testing.T) {
	d, err := New(models.PLC{Driver: models.DriverOPCUA, IPAddress: "127.0.0.1", Port: 4840})
	if err != nil {
		t.Fatalf("expected opcua driver to construct, got error: %v", err)
	}
	if d == nil {
		t.Fatal("expected non-nil driver")
	}
}

func TestNewUnimplementedDrivers(t *testing.T) {
	for _, dt := range []models.DriverType{
		models.DriverModbus, models.DriverS7, models.DriverMitsubishiMC,
		models.DriverFINS, models.DriverEtherNetIP, "bogus",
	} {
		if _, err := New(models.PLC{Driver: dt}); err == nil {
			t.Errorf("expected error constructing driver %q, got nil", dt)
		}
	}
}

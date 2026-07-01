package plc

import (
	"context"

	"pharma-platform/internal/models"
)

// Driver defines the contract implemented by every PLC protocol driver.
type Driver interface {

	// Establish connection to the PLC.
	Connect(ctx context.Context) error

	// Close the connection.
	Close() error

	// Read a single tag.
	Read(ctx context.Context, tag models.Tag) (models.Sample, error)
}

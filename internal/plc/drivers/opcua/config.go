package opcua

import (
	"fmt"
	"time"

	"pharma-platform/internal/models"
)

// Config contains OPC UA client configuration.
type Config struct {
	Endpoint string

	ConnectTimeout time.Duration
	RequestTimeout time.Duration
}

// NewConfig creates an OPC UA configuration from a PLC model.
func NewConfig(plc models.PLC) Config {
	return Config{
		Endpoint: fmt.Sprintf(
			"opc.tcp://%s:%d",
			plc.IPAddress,
			plc.Port,
		),

		ConnectTimeout: 10 * time.Second,
		RequestTimeout: 5 * time.Second,
	}
}

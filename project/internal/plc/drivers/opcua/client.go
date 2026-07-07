package opcua

import (
	"context"
	"fmt"

	gopcua "github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"

	"pharma-platform/internal/models"
	"pharma-platform/internal/plc"
)

// Ensure Client implements the plc.Driver interface.
var _ plc.Driver = (*Client)(nil)

// Client implements the plc.Driver interface using the OPC UA protocol.
//
// A Client represents a single persistent connection to one OPC UA server.
// The connection is established using Connect and reused for subsequent reads.
type Client struct {
	cfg Config

	client *gopcua.Client
}

// New creates a new OPC UA client.
func New(cfg Config) *Client {
	return &Client{
		cfg: cfg,
	}
}

// Connect establishes a connection to the OPC UA server.
func (c *Client) Connect(ctx context.Context) error {
	// Already connected.
	if c.client != nil {
		return nil
	}

	// Apply the configured connection timeout.
	ctx, cancel := context.WithTimeout(ctx, c.cfg.ConnectTimeout)
	defer cancel()

	client, err := gopcua.NewClient(
		c.cfg.Endpoint,
		gopcua.SecurityMode(ua.MessageSecurityModeNone),
		gopcua.SecurityPolicy(ua.SecurityPolicyURINone),
		gopcua.AuthAnonymous(),
	)
	if err != nil {
		return fmt.Errorf("create OPC UA client: %w", err)
	}

	if err := client.Connect(ctx); err != nil {
		return fmt.Errorf("connect to OPC UA server: %w", err)
	}

	c.client = client

	return nil
}

// Close closes the connection to the OPC UA server.
func (c *Client) Close() error {
	panic("not implemented")
}

// Read reads a single tag from the OPC UA server.
func (c *Client) Read(
	ctx context.Context,
	tag models.Tag,
) (models.Sample, error) {
	panic("not implemented")
}

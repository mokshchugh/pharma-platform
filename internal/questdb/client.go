package questdb

import (
	"fmt"
	"net"
	"strconv"
)

type Client struct {
	cfg Config

	conn net.Conn
}

func New(cfg Config) *Client {
	return &Client{
		cfg: cfg,
	}
}

// Connect establishes a TCP connection to QuestDB's ILP endpoint.
func (c *Client) Connect() error {
	if c.conn != nil {
		return nil
	}

	address := net.JoinHostPort(
		c.cfg.Host,
		strconv.Itoa(c.cfg.Port),
	)

	conn, err := net.Dial("tcp", address)
	if err != nil {
		return fmt.Errorf("connect to questdb: %w", err)
	}

	c.conn = conn

	return nil
}

// Close closes the QuestDB connection.
func (c *Client) Close() error {
	if c.conn == nil {
		return nil
	}

	if err := c.conn.Close(); err != nil {
		return fmt.Errorf("close questdb connection: %w", err)
	}

	c.conn = nil

	return nil
}

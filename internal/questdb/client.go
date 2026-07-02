package questdb

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"
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
func (c *Client) Connect(ctx context.Context) error {
	if c.conn != nil {
		return nil
	}

	address := net.JoinHostPort(
		c.cfg.Host,
		strconv.Itoa(c.cfg.Port),
	)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		default:
			conn, err := net.DialTimeout(
				"tcp",
				address,
				5*time.Second,
			)
			if err == nil {
				c.conn = conn
				return nil
			}

			time.Sleep(time.Second)
		}
	}
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

package postgres

import (
	"database/sql"
	"fmt"
	"net"
	"strconv"

	_ "github.com/lib/pq"
)

type Client struct {
	cfg Config

	db *sql.DB
}

func New(cfg Config) *Client {
	return &Client{
		cfg: cfg,
	}
}

func (c *Client) Connect() error {
	if c.db != nil {
		return nil
	}

	host := net.JoinHostPort(
		c.cfg.Host,
		strconv.Itoa(c.cfg.Port),
	)

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=disable",
		c.cfg.User,
		c.cfg.Password,
		host,
		c.cfg.Database,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("open postgres: %w", err)
	}

	db.SetMaxOpenConns(c.cfg.MaxOpenConns)
	db.SetMaxIdleConns(c.cfg.MaxIdleConns)

	if err := db.Ping(); err != nil {
		return fmt.Errorf("ping postgres: %w", err)
	}

	c.db = db

	return nil
}

func (c *Client) Close() error {
	if c.db == nil {
		return nil
	}

	if err := c.db.Close(); err != nil {
		return fmt.Errorf("close postgres: %w", err)
	}

	c.db = nil

	return nil
}

package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

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

func (c *Client) Connect(ctx context.Context) error {
	if c.db != nil {
		return nil
	}

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.cfg.Host,
		c.cfg.Port,
		c.cfg.User,
		c.cfg.Password,
		c.cfg.Database,
	)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		default:
			db, err := sql.Open("postgres", dsn)
			if err == nil {
				if err := db.PingContext(ctx); err == nil {
					db.SetMaxOpenConns(c.cfg.MaxOpenConns)
					db.SetMaxIdleConns(c.cfg.MaxIdleConns)

					c.db = db
					return nil
				}

				_ = db.Close()
			}

			time.Sleep(time.Second)
		}
	}
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

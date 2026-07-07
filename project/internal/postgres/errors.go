package postgres

import "errors"

var (
	ErrNotConnected = errors.New("postgres: client not connected")
)

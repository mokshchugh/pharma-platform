package opcua

import (
	"context"
	"errors"
	"testing"

	"pharma-platform/internal/models"
)

func TestReadBeforeConnect(t *testing.T) {
	c := New(Config{})

	_, err := c.Read(context.Background(), models.Tag{})
	if !errors.Is(err, ErrNotConnected) {
		t.Fatalf("expected ErrNotConnected, got %v", err)
	}
}

func TestCloseBeforeConnect(t *testing.T) {
	c := New(Config{})

	if err := c.Close(); err != nil {
		t.Fatalf("expected nil error closing an unconnected client, got %v", err)
	}
}

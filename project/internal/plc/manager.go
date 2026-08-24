package plc

import (
	"context"
	"fmt"
	"sync"
)

// Manager owns the lifecycle of a set of connected Drivers, keyed by PLC
// ID. It doesn't know how to construct a Driver for a given protocol
// (see internal/plc/registry for that) — it only tracks and connects/
// disconnects drivers handed to it.
type Manager struct {
	mu      sync.RWMutex
	drivers map[string]Driver
}

func NewManager() *Manager {
	return &Manager{
		drivers: make(map[string]Driver),
	}
}

// Add registers a driver for a PLC ID and connects it.
func (m *Manager) Add(ctx context.Context, plcID string, d Driver) error {
	if err := d.Connect(ctx); err != nil {
		return fmt.Errorf("connect plc %s: %w", plcID, err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.drivers[plcID] = d

	return nil
}

// Get returns the driver registered for a PLC ID, if any.
func (m *Manager) Get(plcID string) (Driver, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	d, ok := m.drivers[plcID]
	return d, ok
}

// Remove closes and unregisters the driver for a PLC ID.
func (m *Manager) Remove(plcID string) error {
	m.mu.Lock()
	d, ok := m.drivers[plcID]
	delete(m.drivers, plcID)
	m.mu.Unlock()

	if !ok {
		return nil
	}

	return d.Close()
}

// CloseAll closes every registered driver, collecting any errors.
func (m *Manager) CloseAll() error {
	m.mu.Lock()
	drivers := m.drivers
	m.drivers = make(map[string]Driver)
	m.mu.Unlock()

	var errs []error
	for id, d := range drivers {
		if err := d.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close plc %s: %w", id, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("%d driver(s) failed to close: %v", len(errs), errs)
	}

	return nil
}

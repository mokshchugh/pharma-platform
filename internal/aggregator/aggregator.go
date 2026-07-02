package aggregator

import (
	"context"
	"sync"
	"time"

	"pharma-platform/internal/postgres"
	"pharma-platform/internal/questdb"
)

type Aggregator struct {
	questdb *questdb.Client
	writer  *postgres.Writer

	config Config

	wg sync.WaitGroup
}

func New(
	quest *questdb.Client,
	writer *postgres.Writer,
	config Config,
) *Aggregator {
	return &Aggregator{
		questdb: quest,
		writer:  writer,
		config:  config,
	}
}

func (a *Aggregator) Start(ctx context.Context) error {
	a.wg.Add(1)

	go func() {
		defer a.wg.Done()

		ticker := time.NewTicker(a.config.Interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return

			case <-ticker.C:
				if err := a.aggregate(ctx); err != nil {
					// TODO: logger
				}
			}
		}
	}()

	return nil
}

func (a *Aggregator) Stop() error {
	a.wg.Wait()
	return nil
}

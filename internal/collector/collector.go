package collector

import (
	"context"
	"sync"

	"pharma-platform/internal/models"
	"pharma-platform/internal/plc"
)

type Collector struct {
	driver plc.Driver

	tags []models.Tag

	workQueue chan models.Tag
	samples   chan<- models.Sample

	config Config

	wg sync.WaitGroup
}

func New(
	driver plc.Driver,
	config Config,
	tags []models.Tag,
	samples chan<- models.Sample,
) *Collector {
	return &Collector{
		driver:    driver,
		config:    config,
		tags:      tags,
		workQueue: make(chan models.Tag, config.QueueSize),
		samples:   samples,
	}
}

// Start starts the collector.
func (c *Collector) Start(ctx context.Context) error {
	if err := c.driver.Connect(ctx); err != nil {
		return err
	}

	c.wg.Add(1)
	go c.runScheduler(ctx)

	for i := 0; i < c.config.Workers; i++ {
		c.wg.Add(1)
		go c.runWorker(ctx)
	}

	return nil
}

// Stop stops the collector.
func (c *Collector) Stop() error {
	c.wg.Wait()

	return c.driver.Close()
}

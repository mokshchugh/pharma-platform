package collector

import (
	"context"
)

func (c *Collector) runWorker(ctx context.Context) {
	defer c.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return

		case tag := <-c.workQueue:
			sample, err := c.driver.Read(ctx, tag)
			if err != nil {
				continue
			}

			select {
			case c.samples <- sample:

			case <-ctx.Done():
				return
			}
		}
	}
}

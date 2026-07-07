package collector

import (
	"context"
	"log"
)

func (c *Collector) runWorker(ctx context.Context) {
	defer c.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return

		case tag, ok := <-c.workQueue:
			if !ok {
				return
			}
			sample, err := c.driver.Read(ctx, tag)

			c.mu.Lock()
			delete(c.inFlight, tagKey(tag))
			c.mu.Unlock()

			if err != nil {
				log.Printf(
					"read tag %s/%s: %v",
					tag.PLCID,
					tag.ID,
					err,
				)
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

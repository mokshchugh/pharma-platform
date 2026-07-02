package collector

import (
	"context"
	"time"
)

func (c *Collector) runScheduler(ctx context.Context) {
	defer c.wg.Done()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	lastPoll := make(map[string]time.Time)

	for {
		select {
		case <-ctx.Done():
			return

		case now := <-ticker.C:
			for _, tag := range c.tags {
				if !tag.Enabled {
					continue
				}

				if now.Sub(lastPoll[tag.ID]) < tag.PollInterval {
					continue
				}

				select {
				case c.workQueue <- tag:
					lastPoll[tag.ID] = now

				case <-ctx.Done():
					return
				}
			}
		}
	}
}

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

				key := tagKey(tag)

				if now.Sub(lastPoll[key]) < tag.PollInterval {
					continue
				}

				c.mu.Lock()

				if c.inFlight[key] {
					c.mu.Unlock()
					continue
				}

				c.inFlight[key] = true
				c.mu.Unlock()

				select {
				case c.workQueue <- tag:
					lastPoll[key] = now

				case <-ctx.Done():
					c.mu.Lock()
					delete(c.inFlight, key)
					c.mu.Unlock()
					return
				}
			}
		}
	}
}

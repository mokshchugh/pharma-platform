package collector

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"
	"pharma-platform/internal/config"
	"pharma-platform/internal/models"
)

func BenchmarkCollectorScale(b *testing.B) {
	type benchCase struct {
		name    string
		tags    int
		workers int
		pollMs  int
	}

	cases := []benchCase{
		{"1tag-1worker-100ms", 1, 1, 100},
		{"1tag-16worker-100ms", 1, 16, 100},
		{"10tags-1worker-100ms", 10, 1, 100},
		{"10tags-16worker-100ms", 10, 16, 100},
		{"100tags-16worker-100ms", 100, 16, 100},
		{"1000tags-16worker-100ms", 1000, 16, 100},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			samples := make(chan models.Sample, 100000)
			ccfg := config.CollectorConfig{Workers: tc.workers, QueueSize: 100000}

			tags := make([]models.Tag, tc.tags)
			for i := range tags {
				tags[i] = models.Tag{
					ID:           fmt.Sprintf("tag-%d", i),
					PLCID:        "plc-1",
					Name:         fmt.Sprintf("Tag %d", i),
					Address:      "mock",
					DataType:     models.DataTypeFloat64,
					PollInterval: time.Duration(tc.pollMs) * time.Millisecond,
					Enabled:      true,
				}
			}

			c := New(&mockDriver{}, ccfg, tags, samples)
			ctx, cancel := context.WithCancel(context.Background())

			if err := c.Start(ctx); err != nil {
				b.Fatal(err)
			}

			var count atomic.Int64
			go func() {
				for range samples {
					count.Add(1)
				}
			}()

			time.Sleep(2 * time.Second)
			cancel()
			c.Stop()
			close(samples)
			time.Sleep(50 * time.Millisecond)

			b.ReportMetric(float64(count.Load())/2.0, "s/sec")
		})
	}
}

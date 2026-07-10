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

type mockDriver struct{}

func (m *mockDriver) Connect(ctx context.Context) error { return nil }
func (m *mockDriver) Close() error                     { return nil }
func (m *mockDriver) Read(ctx context.Context, tag models.Tag) (models.Sample, error) {
	return models.Sample{
		Timestamp:   time.Now(),
		MachineID:   fmt.Sprintf("%d", tag.MachineID),
		MachineName: tag.MachineName,
		TagName:     tag.Name,
		Value:       42.0,
		Quality:     models.QualityGood,
	}, nil
}

func BenchmarkCollectorVariants(b *testing.B) {
	type benchCase struct {
		name    string
		tags    int
		workers int
		queue   int
	}

	cases := []benchCase{
		{"1000tags-16workers-10kqueue", 1000, 16, 10000},
		{"1000tags-64workers-10kqueue", 1000, 64, 10000},
		{"1000tags-256workers-10kqueue", 1000, 256, 10000},
		{"500tags-16workers-10kqueue", 500, 16, 10000},
		{"1000tags-16workers-100kqueue", 1000, 16, 100000},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			samples := make(chan models.Sample, 100000)
			ccfg := config.CollectorConfig{Workers: tc.workers, QueueSize: tc.queue}

			tags := make([]models.Tag, tc.tags)
			for i := range tags {
				tags[i] = models.Tag{
					ID:           fmt.Sprintf("tag-%d", i),
					PLCID:        "plc-1",
					Name:         fmt.Sprintf("Tag %d", i),
					Address:      "mock",
					DataType:     models.DataTypeFloat64,
					PollInterval: 100 * time.Millisecond,
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

			b.ReportMetric(float64(count.Load())/2.0, "samples/sec")
		})
	}
}

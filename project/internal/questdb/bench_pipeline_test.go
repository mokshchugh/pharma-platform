package questdb

import (
	"context"
	"testing"
	"time"
	"pharma-platform/internal/models"
)

func BenchmarkFullPipeline(b *testing.B) {
	client := New(Config{
		Host:          "localhost",
		Port:          9009,
		BatchSize:     1000,
		FlushInterval: time.Second,
	})
	
	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		b.Fatal(err)
	}
	
	samples := make(chan models.Sample, 100000)
	writer := NewWriter(client, "plc_samples", samples)
	
	if err := writer.Start(ctx); err != nil {
		b.Fatal(err)
	}
	
	b.ResetTimer()
	
	go func() {
		for i := 0; i < b.N; i++ {
			samples <- models.Sample{
				Timestamp: time.Now(),
				PLCID:     "plc-bench",
				TagID:     "bench-tag",
				Value:     42.0,
				Quality:   models.QualityGood,
			}
		}
		close(samples)
	}()
	
	writer.Stop()
	client.Close()
}

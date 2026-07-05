package questdb

import (
	"testing"
	"time"
	"pharma-platform/internal/models"
)

func BenchmarkEncode(b *testing.B) {
	samples := make([]models.Sample, 1000)
	for i := range samples {
		samples[i] = models.Sample{
			Timestamp: time.Now(),
			PLCID:     "plc-1",
			TagID:     "tag-0",
			Value:     42.0,
			Quality:   models.QualityGood,
		}
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encode("plc_samples", samples)
	}
}

func BenchmarkEncodeParallel(b *testing.B) {
	samples := make([]models.Sample, 1000)
	for i := range samples {
		samples[i] = models.Sample{
			Timestamp: time.Now(),
			PLCID:     "plc-1",
			TagID:     "tag-0",
			Value:     42.0,
			Quality:   models.QualityGood,
		}
	}
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			encode("plc_samples", samples)
		}
	})
}

func BenchmarkFlushEndToEnd(b *testing.B) {
	samples := make([]models.Sample, 1000)
	for i := range samples {
		samples[i] = models.Sample{
			Timestamp: time.Now(),
			PLCID:     "plc-1",
			TagID:     "tag-0",
			Value:     42.0,
			Quality:   models.QualityGood,
		}
	}
	
	data := encode("plc_samples", samples)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = []byte(data)
	}
}

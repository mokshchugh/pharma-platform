package questdb

import "time"

type Config struct {
	Host          string        `yaml:"host"`
	Port          int           `yaml:"port"`
	BatchSize     int           `yaml:"batch_size"`
	FlushInterval time.Duration `yaml:"flush_interval"`
}

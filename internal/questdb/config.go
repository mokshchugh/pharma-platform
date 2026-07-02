package questdb

import "time"

type Config struct {
	Host string
	Port int

	BatchSize     int
	FlushInterval time.Duration
}

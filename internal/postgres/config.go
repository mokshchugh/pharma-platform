package postgres

type Config struct {
	Host string
	Port int

	Database string
	User     string
	Password string

	MaxOpenConns int
	MaxIdleConns int
}

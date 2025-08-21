package constants

import "time"

// Database Configuration Constants
const (
	// Default Database Settings
	DefaultDBHost     = "localhost"
	DefaultDBPort     = "3306"
	DefaultDBUser     = "root"
	DefaultDBPassword = "password"
	DefaultDBName     = "log_analytics"

	// Connection Pool Settings
	DefaultMaxOpenConns    = 25
	DefaultMaxIdleConns    = 5
	DefaultConnMaxLifetime = 5 * time.Minute

	// Environment Variable Keys
	EnvKeyDBHost            = "MYSQL_HOST"
	EnvKeyDBPort            = "MYSQL_PORT"
	EnvKeyDBUser            = "MYSQL_USER"
	EnvKeyDBPassword        = "MYSQL_PASSWORD"
	EnvKeyDBDatabase        = "MYSQL_DATABASE"
	EnvKeyDBMaxOpenConns    = "DB_MAX_OPEN_CONNS"
	EnvKeyDBMaxIdleConns    = "DB_MAX_IDLE_CONNS"
	EnvKeyDBConnMaxLifetime = "DB_CONN_MAX_LIFETIME"
)

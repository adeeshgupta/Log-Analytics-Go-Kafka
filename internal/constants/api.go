package constants

import "time"

// API Configuration Constants
const (
	// Server Configuration
	DefaultServerPort = "8080"

	// Server Timeouts
	DefaultServerReadTimeout  = 30 * time.Second
	DefaultServerWriteTimeout = 30 * time.Second
	DefaultServerIdleTimeout  = 60 * time.Second

	// Environment Variable Keys
	EnvKeyAPIPort            = "API_PORT"
	EnvKeyServerReadTimeout  = "SERVER_READ_TIMEOUT"
	EnvKeyServerWriteTimeout = "SERVER_WRITE_TIMEOUT"
	EnvKeyServerIdleTimeout  = "SERVER_IDLE_TIMEOUT"

	// API Base Paths
	APIPrefix      = "/api"
	APILogsPath    = "/logs"
	APIMetricsPath = "/metrics"
	APIHealthPath  = "/health"
)

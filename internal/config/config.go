package config

import (
	"github.com/adeesh/log-analytics/internal/constants"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	Kafka    KafkaConfig    `json:"kafka"`
	Log      LogConfig      `json:"log"`
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port         string        `json:"port"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout"`
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	Host            string        `json:"host"`
	Port            string        `json:"port"`
	Username        string        `json:"username"`
	Password        string        `json:"password"`
	Database        string        `json:"database"`
	MaxOpenConns    int           `json:"max_open_conns"`
	MaxIdleConns    int           `json:"max_idle_conns"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime"`
}

// KafkaConfig holds Kafka-related configuration
type KafkaConfig struct {
	Brokers          []string `json:"brokers"`
	Topic            string   `json:"topic"`
	GroupID          string   `json:"group_id"`
	AutoOffsetReset  string   `json:"auto_offset_reset"`
	EnableAutoCommit bool     `json:"enable_auto_commit"`
}

// LogConfig holds logging-related configuration
type LogConfig struct {
	Level  string `json:"level"`
	Format string `json:"format"`
}

// Load loads configuration from environment variables
func Load() *Config {
	godotenv.Load()

	config := &Config{
		Server: ServerConfig{
			Port:         getEnv(constants.EnvKeyAPIPort, constants.DefaultServerPort),
			ReadTimeout:  getEnvAsDuration(constants.EnvKeyServerReadTimeout, constants.DefaultServerReadTimeout),
			WriteTimeout: getEnvAsDuration(constants.EnvKeyServerWriteTimeout, constants.DefaultServerWriteTimeout),
			IdleTimeout:  getEnvAsDuration(constants.EnvKeyServerIdleTimeout, constants.DefaultServerIdleTimeout),
		},
		Database: DatabaseConfig{
			Host:            getEnv(constants.EnvKeyDBHost, constants.DefaultDBHost),
			Port:            getEnv(constants.EnvKeyDBPort, constants.DefaultDBPort),
			Username:        getEnv(constants.EnvKeyDBUser, constants.DefaultDBUser),
			Password:        getEnv(constants.EnvKeyDBPassword, constants.DefaultDBPassword),
			Database:        getEnv(constants.EnvKeyDBDatabase, constants.DefaultDBName),
			MaxOpenConns:    getEnvAsInt(constants.EnvKeyDBMaxOpenConns, constants.DefaultMaxOpenConns),
			MaxIdleConns:    getEnvAsInt(constants.EnvKeyDBMaxIdleConns, constants.DefaultMaxIdleConns),
			ConnMaxLifetime: getEnvAsDuration(constants.EnvKeyDBConnMaxLifetime, constants.DefaultConnMaxLifetime),
		},
		Kafka: KafkaConfig{
			Brokers:          getEnvAsSlice(constants.EnvKeyKafkaBrokers, []string{constants.DefaultKafkaBroker}),
			Topic:            getEnv(constants.EnvKeyKafkaTopic, constants.DefaultKafkaTopic),
			GroupID:          getEnv(constants.EnvKeyKafkaGroupID, constants.DefaultConsumerGroupID),
			AutoOffsetReset:  getEnv(constants.EnvKeyKafkaAutoOffsetReset, constants.DefaultAutoOffsetReset),
			EnableAutoCommit: getEnvAsBool(constants.EnvKeyKafkaEnableAutoCommit, true),
		},
		Log: LogConfig{
			Level:  getEnv(constants.EnvKeyLogLevel, constants.DefaultLogLevel),
			Format: getEnv(constants.EnvKeyLogFormat, constants.DefaultLogFormat),
		},
	}

	return config
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getEnvAsSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		// Parse comma-separated values
		values := strings.Split(value, ",")
		// Trim whitespace from each value
		for i, v := range values {
			values[i] = strings.TrimSpace(v)
		}
		return values
	}
	return defaultValue
}

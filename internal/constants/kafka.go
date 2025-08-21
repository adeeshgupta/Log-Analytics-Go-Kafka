package constants

import "time"

// Kafka Configuration Constants
const (
	// Consumer Configuration
	DefaultBatchSize    = 20
	DefaultBatchTimeout = 2 * time.Second

	// Producer Configuration
	DefaultProducerRetryMax = 5

	// Consumer Group Configuration
	DefaultConsumerAutoCommitInterval = 1 * time.Second

	// Kafka Configuration
	DefaultKafkaTopic      = "logs"
	DefaultKafkaBroker     = "localhost:9092"
	DefaultConsumerGroupID = "log-processor-final"
	DefaultAutoOffsetReset = "latest"

	// Environment Variable Keys
	EnvKeyKafkaBrokers          = "KAFKA_BROKERS"
	EnvKeyKafkaTopic            = "KAFKA_TOPIC"
	EnvKeyKafkaGroupID          = "KAFKA_GROUP_ID"
	EnvKeyKafkaAutoOffsetReset  = "KAFKA_AUTO_OFFSET_RESET"
	EnvKeyKafkaEnableAutoCommit = "KAFKA_ENABLE_AUTO_COMMIT"

	// Kafka Headers
	HeaderService   = "service"
	HeaderLevel     = "level"
	HeaderTimestamp = "timestamp"
)

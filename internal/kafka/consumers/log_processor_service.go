package consumers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/adeesh/log-analytics/internal/config"
	"github.com/adeesh/log-analytics/internal/constants"
	"github.com/adeesh/log-analytics/internal/database"
	"github.com/adeesh/log-analytics/internal/database/logs"
	"github.com/adeesh/log-analytics/internal/handlers"
	"github.com/adeesh/log-analytics/internal/models"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IBM/sarama"
)

// LogProcessorService represents the log processing service with integrated batch consumer
type LogProcessorService struct {
	consumer     sarama.ConsumerGroup
	topic        string
	handler      handlers.LogHandler
	logger       *slog.Logger
	batchSize    int
	batchTimeout time.Duration
}

// NewLogProcessorService creates a new log processor service
func NewLogProcessorService(cfg *config.Config, logger *slog.Logger) (*LogProcessorService, error) {
	// Initialize database
	db, err := database.NewGormDB(&cfg.Database)
	if err != nil {
		return nil, err
	}

	// Create log repository
	logRepo := logs.NewLogRepository(db)

	// Create log handlers using the handlers package
	logHandler := handlers.NewLogHandler(logRepo, logger)

	// Create Kafka consumer configuration
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	config.Consumer.Offsets.AutoCommit.Enable = true
	config.Consumer.Offsets.AutoCommit.Interval = constants.DefaultConsumerAutoCommitInterval

	// Set specific version for compatibility
	config.Version = sarama.V3_0_0_0

	// Network configuration
	config.Net.MaxOpenRequests = 5
	config.Net.DialTimeout = 30 * time.Second
	config.Net.ReadTimeout = 30 * time.Second
	config.Net.WriteTimeout = 30 * time.Second

	// Consumer group configuration
	config.Consumer.Group.Session.Timeout = 45 * time.Second
	config.Consumer.Group.Heartbeat.Interval = 10 * time.Second
	config.Consumer.Group.Rebalance.Timeout = 90 * time.Second
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategySticky

	// Create consumer group
	logger.Info("Creating consumer group", "group_id", cfg.Kafka.GroupID, "brokers", cfg.Kafka.Brokers)
	consumer, err := sarama.NewConsumerGroup(cfg.Kafka.Brokers, cfg.Kafka.GroupID, config)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create consumer: %w", err)
	}

	// Test connection and get metadata
	logger.Info("Consumer group created successfully", "group_id", cfg.Kafka.GroupID)

	// Create a test client to verify topic exists
	testClient, err := sarama.NewClient(cfg.Kafka.Brokers, config)
	if err != nil {
		logger.Warn("Failed to create test client", "error", err)
	} else {
		defer testClient.Close()

		// Get topic metadata
		topics, err := testClient.Topics()
		if err != nil {
			logger.Warn("Failed to get topics", "error", err)
		} else {
			logger.Info("Available topics", "topics", topics)
		}

		// Check if our topic exists
		partitions, err := testClient.Partitions(cfg.Kafka.Topic)
		if err != nil {
			logger.Warn("Failed to get topic partitions", "topic", cfg.Kafka.Topic, "error", err)
		} else {
			logger.Info("Topic partitions found", "topic", cfg.Kafka.Topic, "partitions", len(partitions))
		}
	}

	return &LogProcessorService{
		consumer:     consumer,
		topic:        cfg.Kafka.Topic,
		handler:      *logHandler,
		logger:       logger,
		batchSize:    constants.DefaultBatchSize,
		batchTimeout: constants.DefaultBatchTimeout,
	}, nil
}

// Start starts the log processor service
func (s *LogProcessorService) Start(ctx context.Context) error {
	s.logger.Info("Log processor service started",
		"batch_size", s.batchSize,
		"batch_timeout", s.batchTimeout)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		s.logger.Info("Shutdown signal received")
		cancel()
	}()

	// Start consuming messages
	topics := []string{s.topic}
	for {
		err := s.consumer.Consume(ctx, topics, s)
		if err != nil {
			s.logger.Error("Error from consumer", "error", err)
			return err
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
}

// ConsumeClaim implements sarama.ConsumerGroupHandler for batch processing
func (s *LogProcessorService) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	batch := make([]*models.Log, 0, s.batchSize)
	timer := time.NewTimer(s.batchTimeout)
	defer timer.Stop()

	for {
		select {
		case message := <-claim.Messages():
			var log models.Log
			if err := json.Unmarshal(message.Value, &log); err != nil {
				s.logger.Error("Failed to unmarshal log", "error", err)
				session.MarkMessage(message, "")
				continue
			}

			// Add processing metadata
			if log.Timestamp.IsZero() {
				log.Timestamp = time.Now()
			}
			if log.CreatedAt.IsZero() {
				log.CreatedAt = time.Now()
			}

			batch = append(batch, &log)
			session.MarkMessage(message, "")

			// Process batch if it's full
			if len(batch) >= s.batchSize {
				if err := s.processBatch(session.Context(), batch); err != nil {
					s.logger.Error("Failed to process batch", "error", err, "batch_size", len(batch))
				}
				batch = batch[:0]
				timer.Reset(s.batchTimeout)
			}

		case <-timer.C:
			// Process batch on timeout
			if len(batch) > 0 {
				if err := s.processBatch(session.Context(), batch); err != nil {
					s.logger.Error("Failed to process batch on timeout", "error", err, "batch_size", len(batch))
				}
				batch = batch[:0]
			}
			timer.Reset(s.batchTimeout)

		case <-session.Context().Done():
			// Process remaining batch
			if len(batch) > 0 {
				if err := s.processBatch(session.Context(), batch); err != nil {
					s.logger.Error("Failed to process final batch", "error", err, "batch_size", len(batch))
				}
			}
			return nil
		}
	}
}

// Setup implements sarama.ConsumerGroupHandler
func (s *LogProcessorService) Setup(sarama.ConsumerGroupSession) error {
	s.logger.Info("Log processor setup completed")
	return nil
}

// Cleanup implements sarama.ConsumerGroupHandler
func (s *LogProcessorService) Cleanup(sarama.ConsumerGroupSession) error {
	s.logger.Info("Log processor cleanup completed")
	return nil
}

// Close closes the service and its resources
func (s *LogProcessorService) Close() error {
	return s.consumer.Close()
}

// processBatch processes a batch of logs
func (s *LogProcessorService) processBatch(ctx context.Context, logs []*models.Log) error {
	s.logger.Debug("Processing batch", "batch_size", len(logs))
	return s.handler.HandleLogBatch(ctx, logs)
}

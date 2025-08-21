package producers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/adeesh/log-analytics/internal/config"
	"github.com/adeesh/log-analytics/internal/constants"
	"github.com/adeesh/log-analytics/internal/models"
	"log/slog"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
)

// LogCollectorService represents the log collection service with integrated producer
type LogCollectorService struct {
	producer sarama.SyncProducer
	topic    string
	logger   *slog.Logger
}

// NewLogCollectorService creates a new log collector service
func NewLogCollectorService(cfg *config.Config, logger *slog.Logger) (*LogCollectorService, error) {
	// Create Kafka producer configuration
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = constants.DefaultProducerRetryMax
	config.Producer.Return.Successes = true
	config.Producer.Compression = sarama.CompressionSnappy

	// Create producer
	producer, err := sarama.NewSyncProducer(cfg.Kafka.Brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	return &LogCollectorService{
		producer: producer,
		topic:    cfg.Kafka.Topic,
		logger:   logger,
	}, nil
}

// Start starts the log collector service
func (s *LogCollectorService) Start(ctx context.Context) error {
	s.logger.Info("Log collector service started")

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

	// Start generating sample logs
	go s.generateSampleLogs(ctx)

	// Wait for context cancellation
	<-ctx.Done()
	s.logger.Info("Log collector service stopped")
	return nil
}

// Close closes the service and its resources
func (s *LogCollectorService) Close() error {
	return s.producer.Close()
}

// generateSampleLogs generates and sends sample logs to Kafka
func (s *LogCollectorService) generateSampleLogs(ctx context.Context) {
	services := []string{constants.ServiceAPIGateway, constants.ServiceUserService, constants.ServicePaymentService, constants.ServiceOrderService, constants.ServiceNotificationService}
	levels := []models.LogLevel{models.LogLevelDebug, models.LogLevelInfo, models.LogLevelWarn, models.LogLevelError, models.LogLevelFatal}
	methods := []string{constants.MethodGET, constants.MethodPOST, constants.MethodPUT, constants.MethodDELETE}
	paths := []string{constants.PathAPIUsers, constants.PathAPIOrders, constants.PathAPIPayments, constants.PathAPIProducts, constants.PathAPIAuth}

	ticker := time.NewTicker(constants.LogGenerationInterval * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Generate 1-5 logs per second
			count := rand.Intn(constants.MaxLogsPerSecond)

			for i := 0; i < count; i++ {
				log := s.generateRandomLog(services, levels, methods, paths)

				// Send individual log
				if err := s.SendLog(ctx, log); err != nil {
					s.logger.Error("Failed to send log", "error", err)
				}
			}
		}
	}
}

// generateRandomLog creates a random log entry for testing
func (s *LogCollectorService) generateRandomLog(services []string, levels []models.LogLevel, methods []string, paths []string) *models.Log {
	service := services[rand.Intn(len(services))]
	level := levels[rand.Intn(len(levels))]
	method := methods[rand.Intn(len(methods))]
	path := paths[rand.Intn(len(paths))]
	traceID := uuid.New().String()
	userID := fmt.Sprintf(constants.UserIDFormat, rand.Intn(constants.MaxUserID)+1)
	responseTime := rand.Intn(constants.MaxResponseTime-constants.MinResponseTime+1) + constants.MinResponseTime
	responseStatus := constants.StatusOK

	// Generate appropriate message based on level
	var message string
	switch level {
	case models.LogLevelDebug:
		message = fmt.Sprintf(constants.DebugMessageTemplate, method, path)
	case models.LogLevelInfo:
		message = fmt.Sprintf(constants.InfoMessageTemplate, method, path)
	case models.LogLevelWarn:
		message = fmt.Sprintf(constants.WarningMessageTemplate, method, path)
		responseTime = rand.Intn(constants.WarningMaxResponseTime-constants.WarningMinResponseTime+1) + constants.WarningMinResponseTime
	case models.LogLevelError:
		message = fmt.Sprintf(constants.ErrorMessageTemplate, method, path)
		responseStatus = constants.StatusError
		responseTime = rand.Intn(constants.ErrorMaxResponseTime-constants.ErrorMinResponseTime+1) + constants.ErrorMinResponseTime
	case models.LogLevelFatal:
		message = fmt.Sprintf(constants.FatalMessageTemplate, service)
		responseStatus = constants.StatusError
		responseTime = rand.Intn(constants.FatalMaxResponseTime-constants.FatalMinResponseTime+1) + constants.FatalMinResponseTime
	}

	// Add some error messages for variety
	if level == models.LogLevelError || level == models.LogLevelFatal {
		errorMessages := []string{
			constants.ErrorDatabaseConnection,
			constants.ErrorExternalTimeout,
			constants.ErrorInvalidPayload,
			constants.ErrorAuthentication,
			constants.ErrorResourceNotFound,
			constants.ErrorInternalServer,
			constants.ErrorRateLimit,
		}
		message = errorMessages[rand.Intn(len(errorMessages))]
	}

	return &models.Log{
		Timestamp:      time.Now(),
		Level:          level,
		Service:        service,
		Message:        message,
		TraceID:        &traceID,
		UserID:         &userID,
		RequestMethod:  &method,
		RequestPath:    &path,
		ResponseStatus: &responseStatus,
		ResponseTimeMs: &responseTime,
		CreatedAt:      time.Now(),
	}
}

// SendLog sends a log message to Kafka
func (s *LogCollectorService) SendLog(_ context.Context, log *models.Log) error {

	// Generate message ID if not present
	if log.TraceID == nil {
		traceID := uuid.New().String()
		log.TraceID = &traceID
	}

	// Serialize log to JSON
	value, err := json.Marshal(log)
	if err != nil {
		return fmt.Errorf("failed to marshal log: %w", err)
	}

	// Create Kafka message
	message := &sarama.ProducerMessage{
		Topic: s.topic,
		Key:   sarama.StringEncoder(*log.TraceID),
		Value: sarama.ByteEncoder(value),
		Headers: []sarama.RecordHeader{
			{Key: []byte(constants.HeaderService), Value: []byte(log.Service)},
			{Key: []byte(constants.HeaderLevel), Value: []byte(string(log.Level))},
			{Key: []byte(constants.HeaderTimestamp), Value: []byte(log.Timestamp.Format(time.RFC3339))},
		},
	}

	// Send message
	partition, offset, err := s.producer.SendMessage(message)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	s.logger.Debug("Log sent", "topic", s.topic, "partition", partition, "offset", offset)
	return nil
}

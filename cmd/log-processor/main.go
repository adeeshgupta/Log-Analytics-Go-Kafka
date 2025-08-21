package main

import (
	"context"
	"github.com/adeesh/log-analytics/internal/config"
	"github.com/adeesh/log-analytics/internal/kafka/consumers"
	"log/slog"
	"os"
)

func main() {
	// Initialize logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Load configuration
	cfg := config.Load()

	// Create log processor service
	service, err := consumers.NewLogProcessorService(cfg, logger)
	if err != nil {
		logger.Error("Failed to create log processor service", "error", err)
		os.Exit(1)
	}
	defer service.Close()

	// Start the service
	if err := service.Start(context.Background()); err != nil {
		logger.Error("Log processor service error", "error", err)
		os.Exit(1)
	}
}

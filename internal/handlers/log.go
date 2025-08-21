package handlers

import (
	"context"
	"github.com/adeesh/log-analytics/internal/database/logs"
	"github.com/adeesh/log-analytics/internal/models"
	"net/http"
	"strconv"
	"time"

	"log/slog"

	"github.com/gin-gonic/gin"
)

// LogHandler handles log-related HTTP requests
type LogHandler struct {
	logRepo logs.LogRepository
	logger  *slog.Logger
}

// NewLogHandler creates a new log handler
func NewLogHandler(logRepo logs.LogRepository, logger *slog.Logger) *LogHandler {
	return &LogHandler{
		logRepo: logRepo,
		logger:  logger,
	}
}

// GetLogs retrieves logs based on query parameters
func (h *LogHandler) GetLogs(c *gin.Context) {
	// Parse query parameters
	filter := &models.LogFilter{}

	if level := c.Query("level"); level != "" {
		logLevel := models.LogLevel(level)
		filter.Level = &logLevel
	}

	if service := c.Query("service"); service != "" {
		filter.Service = &service
	}

	if traceID := c.Query("trace_id"); traceID != "" {
		filter.TraceID = &traceID
	}

	if userID := c.Query("user_id"); userID != "" {
		filter.UserID = &userID
	}

	if startTime := c.Query("start_time"); startTime != "" {
		if t, err := time.Parse(time.RFC3339, startTime); err == nil {
			filter.StartTime = &t
		}
	}

	if endTime := c.Query("end_time"); endTime != "" {
		if t, err := time.Parse(time.RFC3339, endTime); err == nil {
			filter.EndTime = &t
		}
	}

	if search := c.Query("search"); search != "" {
		filter.Search = &search
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filter.Limit = limit
		}
	} else {
		filter.Limit = 100 // default limit
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filter.Offset = offset
		}
	}

	// Get logs from database
	responseLogs, err := h.logRepo.GetLogs(c.Request.Context(), filter)
	if err != nil {
		h.logger.Error("Failed to get logs", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":   responseLogs,
		"count":  len(responseLogs),
		"filter": filter,
	})
}

// GetLogsByTraceID retrieves all logs for a specific trace ID
func (h *LogHandler) GetLogsByTraceID(c *gin.Context) {
	traceID := c.Param("traceID")
	if traceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Trace ID is required"})
		return
	}

	responseLogs, err := h.logRepo.GetLogsByTraceID(c.Request.Context(), traceID)
	if err != nil {
		h.logger.Error("Failed to get logs by trace ID", "error", err, "trace_id", traceID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"trace_id": traceID,
		"logs":     responseLogs,
		"count":    len(responseLogs),
	})
}

// GetMetrics retrieves system metrics and statistics
func (h *LogHandler) GetMetrics(c *gin.Context) {
	// Parse time range with defaults
	endTime := time.Now()
	startTime := endTime.Add(-24 * time.Hour) // Default to last 24 hours

	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			startTime = t
		}
	}

	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			endTime = t
		}
	}

	// Get stats from database
	stats, err := h.logRepo.GetLogStats(c.Request.Context(), startTime, endTime)
	if err != nil {
		h.logger.Error("Failed to get metrics", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve metrics"})
		return
	}

	// Calculate additional metrics
	totalRequests := stats.TotalLogs
	errorRate := 0.0
	if totalRequests > 0 {
		errorRate = float64(stats.ErrorCount+stats.FatalCount) / float64(totalRequests) * 100
	}

	// Calculate time duration for requests per minute
	duration := endTime.Sub(startTime)
	minutes := duration.Minutes()
	if minutes <= 0 {
		minutes = 1 // Avoid division by zero
	}

	response := gin.H{
		// Raw statistics
		"stats": gin.H{
			"total_logs":        stats.TotalLogs,
			"error_count":       stats.ErrorCount,
			"warning_count":     stats.WarningCount,
			"info_count":        stats.InfoCount,
			"debug_count":       stats.DebugCount,
			"fatal_count":       stats.FatalCount,
			"avg_response_time": stats.AvgResponseTime,
			"top_services":      stats.TopServices,
			"top_errors":        stats.TopErrors,
			"time_series":       stats.TimeSeries,
		},
		// Calculated metrics
		"metrics": gin.H{
			"total_requests":      totalRequests,
			"error_count":         stats.ErrorCount + stats.FatalCount,
			"error_rate_percent":  errorRate,
			"avg_response_time":   stats.AvgResponseTime,
			"requests_per_minute": float64(totalRequests) / minutes,
		},
		// Time range information
		"time_range": gin.H{
			"start_time":       startTime,
			"end_time":         endTime,
			"duration_minutes": minutes,
		},
		"timestamp": time.Now(),
	}

	c.JSON(http.StatusOK, response)
}

// HandleLog processes a single log message from Kafka
func (h *LogHandler) HandleLog(ctx context.Context, log *models.Log) error {
	// Store log in database
	if err := h.logRepo.CreateLog(ctx, log); err != nil {
		h.logger.Error("Failed to store log",
			"error", err,
			"trace_id", log.TraceID,
			"service", log.Service)
		return err
	}

	h.logger.Info("Log processed successfully",
		"trace_id", log.TraceID,
		"service", log.Service,
		"level", log.Level,
		"message", log.Message)

	return nil
}

// HandleLogBatch processes a batch of log messages from Kafka
func (h *LogHandler) HandleLogBatch(ctx context.Context, logs []*models.Log) error {
	// Store logs in database
	if err := h.logRepo.CreateLogBatch(ctx, logs); err != nil {
		h.logger.Error("Failed to store log batch",
			"error", err,
			"batch_size", len(logs))
		return err
	}

	h.logger.Info("Log batch processed successfully",
		"batch_size", len(logs),
		"services", getUniqueServices(logs))

	return nil
}

// getUniqueServices extracts unique service names from a batch of logs
func getUniqueServices(logs []*models.Log) []string {
	services := make(map[string]bool)
	for _, log := range logs {
		services[log.Service] = true
	}

	result := make([]string, 0, len(services))
	for service := range services {
		result = append(result, service)
	}
	return result
}

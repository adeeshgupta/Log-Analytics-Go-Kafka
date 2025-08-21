package logs

import (
	"context"
	"fmt"
	"github.com/adeesh/log-analytics/internal/database"
	"github.com/adeesh/log-analytics/internal/models"
	"time"
)

// GormLogRepository represents log-related database operations using GORM
type GormLogRepository struct {
	db *database.GormDB
}

// LogRepository defines the interface for log-related database operations
type LogRepository interface {
	// CreateLog inserts a new log entry
	CreateLog(ctx context.Context, log *models.Log) error
	// CreateLogBatch inserts multiple log entries
	CreateLogBatch(ctx context.Context, logs []*models.Log) error
	// GetLogs retrieves logs based on filters
	GetLogs(ctx context.Context, filter *models.LogFilter) ([]*models.Log, error)
	// GetLogStats retrieves aggregated log statistics
	GetLogStats(ctx context.Context, startTime, endTime time.Time) (*models.LogStats, error)
	// GetLogsByTraceID retrieves all logs for a specific trace ID
	GetLogsByTraceID(ctx context.Context, traceID string) ([]*models.Log, error)
}

// NewLogRepository creates a new log repository
func NewLogRepository(db *database.GormDB) LogRepository {
	return &GormLogRepository{db: db}
}

// CreateLog inserts a new log entry
func (r *GormLogRepository) CreateLog(ctx context.Context, log *models.Log) error {
	result := r.db.GetDB().WithContext(ctx).Create(log)
	if result.Error != nil {
		return fmt.Errorf("failed to create log: %w", result.Error)
	}
	return nil
}

// CreateLogBatch inserts multiple log entries
func (r *GormLogRepository) CreateLogBatch(ctx context.Context, logs []*models.Log) error {
	if len(logs) == 0 {
		return nil
	}
	result := r.db.GetDB().WithContext(ctx).CreateInBatches(logs, 100)
	if result.Error != nil {
		return fmt.Errorf("failed to create log batch: %w", result.Error)
	}
	return nil
}

// GetLogs retrieves logs based on filters
func (r *GormLogRepository) GetLogs(ctx context.Context, filter *models.LogFilter) ([]*models.Log, error) {
	query := r.db.GetDB().WithContext(ctx).Model(&models.Log{})
	if filter.Level != nil {
		query = query.Where("level = ?", *filter.Level)
	}
	if filter.Service != nil {
		query = query.Where("service = ?", *filter.Service)
	}
	if filter.TraceID != nil {
		query = query.Where("trace_id = ?", *filter.TraceID)
	}
	if filter.UserID != nil {
		query = query.Where("user_id = ?", *filter.UserID)
	}
	if filter.StartTime != nil {
		query = query.Where("timestamp >= ?", *filter.StartTime)
	}
	if filter.EndTime != nil {
		query = query.Where("timestamp <= ?", *filter.EndTime)
	}
	if filter.Search != nil {
		query = query.Where("MATCH(message) AGAINST(? IN BOOLEAN MODE)", *filter.Search)
	}
	query = query.Order("timestamp DESC")
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}
	var logs []*models.Log
	if err := query.Find(&logs).Error; err != nil {
		return nil, fmt.Errorf("failed to get logs: %w", err)
	}
	return logs, nil
}

// GetLogStats retrieves aggregated log statistics
func (r *GormLogRepository) GetLogStats(ctx context.Context, startTime, endTime time.Time) (*models.LogStats, error) {
	stats := &models.LogStats{}

	// Get total counts by level
	var result struct {
		TotalLogs       int64   `json:"total_logs"`
		ErrorCount      int64   `json:"error_count"`
		WarningCount    int64   `json:"warning_count"`
		InfoCount       int64   `json:"info_count"`
		DebugCount      int64   `json:"debug_count"`
		FatalCount      int64   `json:"fatal_count"`
		AvgResponseTime float64 `json:"avg_response_time"`
	}

	err := r.db.GetDB().WithContext(ctx).Model(&models.Log{}).
		Select(`
			COUNT(*) as total_logs,
			SUM(CASE WHEN level = 'ERROR' THEN 1 ELSE 0 END) as error_count,
			SUM(CASE WHEN level = 'WARN' THEN 1 ELSE 0 END) as warning_count,
			SUM(CASE WHEN level = 'INFO' THEN 1 ELSE 0 END) as info_count,
			SUM(CASE WHEN level = 'DEBUG' THEN 1 ELSE 0 END) as debug_count,
			SUM(CASE WHEN level = 'FATAL' THEN 1 ELSE 0 END) as fatal_count,
			AVG(response_time_ms) as avg_response_time
		`).
		Where("timestamp BETWEEN ? AND ?", startTime, endTime).
		Scan(&result).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get log stats: %w", err)
	}

	stats.TotalLogs = result.TotalLogs
	stats.ErrorCount = result.ErrorCount
	stats.WarningCount = result.WarningCount
	stats.InfoCount = result.InfoCount
	stats.DebugCount = result.DebugCount
	stats.FatalCount = result.FatalCount
	stats.AvgResponseTime = result.AvgResponseTime

	// Get top services
	var serviceCounts []models.ServiceCount
	err = r.db.GetDB().WithContext(ctx).Model(&models.Log{}).
		Select("service, COUNT(*) as count").
		Where("timestamp BETWEEN ? AND ?", startTime, endTime).
		Group("service").
		Order("count DESC").
		Limit(10).
		Scan(&serviceCounts).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get service stats: %w", err)
	}
	stats.TopServices = serviceCounts

	// Get top errors
	var errorCounts []models.ErrorCount
	err = r.db.GetDB().WithContext(ctx).Model(&models.Log{}).
		Select("message, COUNT(*) as count").
		Where("timestamp BETWEEN ? AND ? AND level IN (?, ?)", startTime, endTime, "ERROR", "FATAL").
		Group("message").
		Order("count DESC").
		Limit(10).
		Scan(&errorCounts).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get error stats: %w", err)
	}
	stats.TopErrors = errorCounts

	return stats, nil
}

// GetLogsByTraceID retrieves all logs for a specific trace ID
func (r *GormLogRepository) GetLogsByTraceID(ctx context.Context, traceID string) ([]*models.Log, error) {
	var logs []*models.Log
	err := r.db.GetDB().WithContext(ctx).
		Where("trace_id = ?", traceID).
		Order("timestamp ASC").
		Find(&logs).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get logs by trace ID: %w", err)
	}
	return logs, nil
}

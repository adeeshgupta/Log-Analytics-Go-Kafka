package models

import (
	"time"
)

// LogLevel represents the severity level of a log entry
type LogLevel string

const (
	LogLevelDebug LogLevel = "DEBUG"
	LogLevelInfo  LogLevel = "INFO"
	LogLevelWarn  LogLevel = "WARN"
	LogLevelError LogLevel = "ERROR"
	LogLevelFatal LogLevel = "FATAL"
)

// Log represents a log entry in the system
type Log struct {
	ID             uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Timestamp      time.Time `json:"timestamp" gorm:"index;not null"`
	Level          LogLevel  `json:"level" gorm:"type:enum('DEBUG','INFO','WARN','ERROR','FATAL');index;not null" validate:"required,oneof=DEBUG INFO WARN ERROR FATAL"`
	Service        string    `json:"service" gorm:"index;not null;size:100" validate:"required"`
	Message        string    `json:"message" gorm:"type:text;not null" validate:"required"`
	TraceID        *string   `json:"trace_id,omitempty" gorm:"index;size:50"`
	UserID         *string   `json:"user_id,omitempty" gorm:"index;size:50"`
	RequestMethod  *string   `json:"request_method,omitempty" gorm:"size:10"`
	RequestPath    *string   `json:"request_path,omitempty" gorm:"size:500"`
	ResponseStatus *int      `json:"response_status,omitempty"`
	ResponseTimeMs *int      `json:"response_time_ms,omitempty"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// LogFilter represents filters for querying logs
type LogFilter struct {
	Level     *LogLevel  `json:"level,omitempty"`
	Service   *string    `json:"service,omitempty"`
	TraceID   *string    `json:"trace_id,omitempty"`
	UserID    *string    `json:"user_id,omitempty"`
	StartTime *time.Time `json:"start_time,omitempty"`
	EndTime   *time.Time `json:"end_time,omitempty"`
	Search    *string    `json:"search,omitempty"`
	Limit     int        `json:"limit,omitempty"`
	Offset    int        `json:"offset,omitempty"`
}

// LogStats represents aggregated statistics for logs
type LogStats struct {
	TotalLogs       int64            `json:"total_logs"`
	ErrorCount      int64            `json:"error_count"`
	WarningCount    int64            `json:"warning_count"`
	InfoCount       int64            `json:"info_count"`
	DebugCount      int64            `json:"debug_count"`
	FatalCount      int64            `json:"fatal_count"`
	AvgResponseTime float64          `json:"avg_response_time"`
	TopServices     []ServiceCount   `json:"top_services"`
	TopErrors       []ErrorCount     `json:"top_errors"`
	TimeSeries      []TimeSeriesData `json:"time_series"`
}

// ServiceCount represents service log count
type ServiceCount struct {
	Service string `json:"service"`
	Count   int64  `json:"count"`
}

// ErrorCount represents error message count
type ErrorCount struct {
	Message string `json:"message"`
	Count   int64  `json:"count"`
}

// TimeSeriesData represents time series data point
type TimeSeriesData struct {
	Timestamp time.Time `json:"timestamp"`
	Count     int64     `json:"count"`
	ErrorRate float64   `json:"error_rate"`
}

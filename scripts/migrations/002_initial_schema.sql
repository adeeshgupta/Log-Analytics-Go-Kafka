-- Initial Database Schema Migration
-- This script creates all tables for the Log Analytics System

-- Create logs table
CREATE TABLE IF NOT EXISTS logs (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    timestamp DATETIME NOT NULL,
    level ENUM('debug', 'info', 'warn', 'error', 'fatal') NOT NULL,
    service VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    trace_id VARCHAR(255),
    user_id VARCHAR(255),
    request_method VARCHAR(10),
    request_path VARCHAR(500),
    response_status INT,
    response_time_ms INT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    -- Indexes for better query performance
    INDEX idx_timestamp (timestamp),
    INDEX idx_level (level),
    INDEX idx_service (service),
    INDEX idx_trace_id (trace_id),
    INDEX idx_user_id (user_id),
    INDEX idx_created_at (created_at),
    INDEX idx_level_service (level, service),
    INDEX idx_timestamp_level (timestamp, level),
    INDEX idx_service_level_created (service, level, created_at),
    INDEX idx_timestamp_level_service (timestamp, level, service)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create alert_rules table
CREATE TABLE IF NOT EXISTS alert_rules (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    `condition` TEXT NOT NULL COMMENT 'SQL condition for the alert',
    threshold DOUBLE NOT NULL,
    time_window INT NOT NULL COMMENT 'Time window in minutes',
    severity ENUM('low', 'medium', 'high', 'critical') NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    -- Indexes
    INDEX idx_enabled (enabled),
    INDEX idx_severity (severity),
    INDEX idx_created_at (created_at),
    INDEX idx_enabled_severity (enabled, severity)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create alerts table
CREATE TABLE IF NOT EXISTS alerts (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    rule_id BIGINT UNSIGNED NOT NULL,
    message TEXT NOT NULL,
    severity ENUM('low', 'medium', 'high', 'critical') NOT NULL,
    value DOUBLE NOT NULL COMMENT 'Actual value that triggered the alert',
    status ENUM('active', 'resolved', 'acknowledged') NOT NULL DEFAULT 'active',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    resolved_at DATETIME NULL,
    acknowledged_at DATETIME NULL,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    -- Foreign key constraint
    FOREIGN KEY (rule_id) REFERENCES alert_rules(id) ON DELETE CASCADE,
    
    -- Indexes
    INDEX idx_rule_id (rule_id),
    INDEX idx_status (status),
    INDEX idx_severity (severity),
    INDEX idx_created_at (created_at),
    INDEX idx_status_created (status, created_at),
    INDEX idx_rule_status (rule_id, status),
    INDEX idx_status_severity_created (status, severity, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Migration completed successfully 

package models

import (
	"time"
)

// AlertRule represents an alert rule configuration
type AlertRule struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description"`
	Condition   string    `json:"condition" gorm:"not null"` // SQL condition for the alert
	Threshold   float64   `json:"threshold" gorm:"not null"`
	TimeWindow  int       `json:"time_window" gorm:"not null"` // in minutes
	Severity    string    `json:"severity" gorm:"type:enum('low','medium','high','critical');not null"`    // low, medium, high, critical
	Enabled     bool      `json:"enabled" gorm:"default:true"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
} 

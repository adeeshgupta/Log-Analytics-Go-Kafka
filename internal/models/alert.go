package models

import (
	"time"
)

// Alert represents a triggered alert
type Alert struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	RuleID         uint      `json:"rule_id" gorm:"not null"`
	Rule           AlertRule `json:"rule" gorm:"foreignKey:RuleID"`
	Message        string    `json:"message" gorm:"not null"`
	Severity       string    `json:"severity" gorm:"type:enum('low','medium','high','critical');not null"`
	Value          float64   `json:"value" gorm:"not null"` // actual value that triggered the alert
	Status         string    `json:"status" gorm:"type:enum('active','resolved','acknowledged');default:'active'"` // active, resolved, acknowledged
	CreatedAt      time.Time `json:"created_at"`
	ResolvedAt     *time.Time `json:"resolved_at"`
	AcknowledgedAt *time.Time `json:"acknowledged_at"`
}

// AlertStats represents alert statistics
type AlertStats struct {
	TotalAlerts    int64 `json:"total_alerts"`
	ActiveAlerts   int64 `json:"active_alerts"`
	ResolvedAlerts int64 `json:"resolved_alerts"`
	CriticalAlerts int64 `json:"critical_alerts"`
	HighAlerts     int64 `json:"high_alerts"`
	MediumAlerts   int64 `json:"medium_alerts"`
	LowAlerts      int64 `json:"low_alerts"`
}

// AlertFilter represents filters for querying alerts
type AlertFilter struct {
	Status   *string    `json:"status"`
	Severity *string    `json:"severity"`
	RuleID   *uint      `json:"rule_id"`
	From     *time.Time `json:"from"`
	To       *time.Time `json:"to"`
	Limit    *int       `json:"limit"`
	Offset   *int       `json:"offset"`
} 

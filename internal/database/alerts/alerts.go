package alerts

import (
	"context"
	"github.com/adeesh/log-analytics/internal/models"
	"time"

	"gorm.io/gorm"
)

// AlertRepository defines the interface for alert operations
type AlertRepository interface {
	CreateAlert(ctx context.Context, alert *models.Alert) error
	GetAlerts(ctx context.Context, filter *models.AlertFilter) ([]models.Alert, error)
	GetAlertByID(ctx context.Context, id uint) (*models.Alert, error)
	UpdateAlert(ctx context.Context, alert *models.Alert) error
	GetAlertStats(ctx context.Context) (*models.AlertStats, error)
	GetActiveAlerts(ctx context.Context) ([]models.Alert, error)
	ResolveAlert(ctx context.Context, id uint) error
	AcknowledgeAlert(ctx context.Context, id uint) error
}

// GormAlertRepository implements AlertRepository using GORM
type GormAlertRepository struct {
	db *gorm.DB
}

// NewAlertRepository creates a new alert repository
func NewAlertRepository(db *gorm.DB) AlertRepository {
	return &GormAlertRepository{db: db}
}

// CreateAlert creates a new alert
func (r *GormAlertRepository) CreateAlert(ctx context.Context, alert *models.Alert) error {
	return r.db.WithContext(ctx).Create(alert).Error
}

// GetAlerts retrieves alerts with filters
func (r *GormAlertRepository) GetAlerts(ctx context.Context, filter *models.AlertFilter) ([]models.Alert, error) {
	query := r.db.WithContext(ctx).Preload("Rule")

	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	if filter.Severity != nil {
		query = query.Where("severity = ?", *filter.Severity)
	}
	if filter.RuleID != nil {
		query = query.Where("rule_id = ?", *filter.RuleID)
	}
	if filter.From != nil {
		query = query.Where("created_at >= ?", *filter.From)
	}
	if filter.To != nil {
		query = query.Where("created_at <= ?", *filter.To)
	}

	// Apply pagination
	if filter.Limit != nil {
		query = query.Limit(*filter.Limit)
	}
	if filter.Offset != nil {
		query = query.Offset(*filter.Offset)
	}

	var alerts []models.Alert
	err := query.Order("created_at DESC").Find(&alerts).Error
	return alerts, err
}

// GetAlertByID retrieves an alert by ID
func (r *GormAlertRepository) GetAlertByID(ctx context.Context, id uint) (*models.Alert, error) {
	var alert models.Alert
	err := r.db.WithContext(ctx).Preload("Rule").First(&alert, id).Error
	if err != nil {
		return nil, err
	}
	return &alert, nil
}

// UpdateAlert updates an alert
func (r *GormAlertRepository) UpdateAlert(ctx context.Context, alert *models.Alert) error {
	return r.db.WithContext(ctx).Save(alert).Error
}

// GetAlertStats retrieves alert statistics
func (r *GormAlertRepository) GetAlertStats(ctx context.Context) (*models.AlertStats, error) {
	var stats models.AlertStats

	// Total alerts
	if err := r.db.WithContext(ctx).Model(&models.Alert{}).Count(&stats.TotalAlerts).Error; err != nil {
		return nil, err
	}

	// Active alerts
	if err := r.db.WithContext(ctx).Model(&models.Alert{}).Where("status = ?", "active").Count(&stats.ActiveAlerts).Error; err != nil {
		return nil, err
	}

	// Resolved alerts
	if err := r.db.WithContext(ctx).Model(&models.Alert{}).Where("status = ?", "resolved").Count(&stats.ResolvedAlerts).Error; err != nil {
		return nil, err
	}

	// Alerts by severity
	if err := r.db.WithContext(ctx).Model(&models.Alert{}).Where("severity = ?", "critical").Count(&stats.CriticalAlerts).Error; err != nil {
		return nil, err
	}
	if err := r.db.WithContext(ctx).Model(&models.Alert{}).Where("severity = ?", "high").Count(&stats.HighAlerts).Error; err != nil {
		return nil, err
	}
	if err := r.db.WithContext(ctx).Model(&models.Alert{}).Where("severity = ?", "medium").Count(&stats.MediumAlerts).Error; err != nil {
		return nil, err
	}
	if err := r.db.WithContext(ctx).Model(&models.Alert{}).Where("severity = ?", "low").Count(&stats.LowAlerts).Error; err != nil {
		return nil, err
	}

	return &stats, nil
}

// GetActiveAlerts retrieves all active alerts
func (r *GormAlertRepository) GetActiveAlerts(ctx context.Context) ([]models.Alert, error) {
	var alerts []models.Alert
	err := r.db.WithContext(ctx).Preload("Rule").Where("status = ?", "active").Order("created_at DESC").Find(&alerts).Error
	return alerts, err
}

// ResolveAlert resolves an alert
func (r *GormAlertRepository) ResolveAlert(ctx context.Context, id uint) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&models.Alert{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":      "resolved",
		"resolved_at": &now,
		"updated_at":  now,
	}).Error
}

// AcknowledgeAlert acknowledges an alert
func (r *GormAlertRepository) AcknowledgeAlert(ctx context.Context, id uint) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&models.Alert{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":          "acknowledged",
		"acknowledged_at": &now,
		"updated_at":      now,
	}).Error
}

package alert_rules

import (
	"context"
	"github.com/adeesh/log-analytics/internal/models"

	"gorm.io/gorm"
)

// AlertRuleRepository defines the interface for alert rule operations
type AlertRuleRepository interface {
	CreateAlertRule(ctx context.Context, rule *models.AlertRule) error
	GetAlertRules(ctx context.Context) ([]models.AlertRule, error)
	GetAlertRuleByID(ctx context.Context, id uint) (*models.AlertRule, error)
	UpdateAlertRule(ctx context.Context, rule *models.AlertRule) error
	DeleteAlertRule(ctx context.Context, id uint) error
}

// GormAlertRuleRepository implements AlertRuleRepository using GORM
type GormAlertRuleRepository struct {
	db *gorm.DB
}

// NewAlertRuleRepository creates a new alert rule repository
func NewAlertRuleRepository(db *gorm.DB) AlertRuleRepository {
	return &GormAlertRuleRepository{db: db}
}

// CreateAlertRule creates a new alert rule
func (r *GormAlertRuleRepository) CreateAlertRule(ctx context.Context, rule *models.AlertRule) error {
	return r.db.WithContext(ctx).Create(rule).Error
}

// GetAlertRules retrieves all alert rules
func (r *GormAlertRuleRepository) GetAlertRules(ctx context.Context) ([]models.AlertRule, error) {
	var rules []models.AlertRule
	err := r.db.WithContext(ctx).Find(&rules).Error
	return rules, err
}

// GetAlertRuleByID retrieves an alert rule by ID
func (r *GormAlertRuleRepository) GetAlertRuleByID(ctx context.Context, id uint) (*models.AlertRule, error) {
	var rule models.AlertRule
	err := r.db.WithContext(ctx).First(&rule, id).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

// UpdateAlertRule updates an alert rule
func (r *GormAlertRuleRepository) UpdateAlertRule(ctx context.Context, rule *models.AlertRule) error {
	return r.db.WithContext(ctx).Save(rule).Error
}

// DeleteAlertRule deletes an alert rule
func (r *GormAlertRuleRepository) DeleteAlertRule(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.AlertRule{}, id).Error
}

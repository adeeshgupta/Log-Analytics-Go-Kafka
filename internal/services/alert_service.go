package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/adeesh/log-analytics/internal/database/alert_rules"
	"github.com/adeesh/log-analytics/internal/database/alerts"
	"github.com/adeesh/log-analytics/internal/models"
	"log/slog"
	"time"
)

// AlertService handles alert rule evaluation and alert creation
type AlertService struct {
	alertRuleRepo alert_rules.AlertRuleRepository
	alertRepo     alerts.AlertRepository
	db            *sql.DB
	logger        *slog.Logger
}

// NewAlertService creates a new alert service
func NewAlertService(alertRuleRepo alert_rules.AlertRuleRepository, alertRepo alerts.AlertRepository, db *sql.DB, logger *slog.Logger) *AlertService {
	return &AlertService{
		alertRuleRepo: alertRuleRepo,
		alertRepo:     alertRepo,
		db:            db,
		logger:        logger,
	}
}

// StartAlertChecker starts the background alert checker
func (s *AlertService) StartAlertChecker(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	s.logger.Info("Alert checker started", "interval", interval)

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Alert checker stopped")
			return
		case <-ticker.C:
			if err := s.CheckAlertRules(ctx); err != nil {
				s.logger.Error("Failed to check alert rules", "error", err)
			}
		}
	}
}

// CheckAlertRules evaluates all enabled alert rules and creates alerts if conditions are met
func (s *AlertService) CheckAlertRules(ctx context.Context) error {
	// Get all enabled alert rules
	rules, err := s.alertRuleRepo.GetAlertRules(ctx)
	if err != nil {
		return fmt.Errorf("failed to get alert rules: %w", err)
	}

	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}

		if err := s.evaluateRule(ctx, &rule); err != nil {
			s.logger.Error("Failed to evaluate alert rule", "error", err, "rule_id", rule.ID, "rule_name", rule.Name)
		}
	}

	return nil
}

// evaluateRule evaluates a single alert rule
func (s *AlertService) evaluateRule(ctx context.Context, rule *models.AlertRule) error {
	// Build the SQL query based on the rule condition
	query := s.buildQuery(rule)

	// Execute the query
	var result float64
	err := s.db.QueryRowContext(ctx, query).Scan(&result)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// No data found, which means no alert should be triggered
			return nil
		}
		return fmt.Errorf("failed to execute alert query: %w", err)
	}

	// Check if the result exceeds the threshold
	if result >= rule.Threshold {
		// Check if there's already an active alert for this rule
		activeAlerts, err := s.alertRepo.GetAlerts(ctx, &models.AlertFilter{
			RuleID: &rule.ID,
			Status: func() *string { s := "active"; return &s }(),
		})
		if err != nil {
			return fmt.Errorf("failed to check existing alerts: %w", err)
		}

		// If no active alert exists, create a new one
		if len(activeAlerts) == 0 {
			alert := &models.Alert{
				RuleID:    rule.ID,
				Message:   fmt.Sprintf("Alert rule '%s' triggered: %s = %.2f (threshold: %.2f)", rule.Name, rule.Condition, result, rule.Threshold),
				Severity:  rule.Severity,
				Value:     result,
				Status:    "active",
				CreatedAt: time.Now(),
			}

			if err := s.alertRepo.CreateAlert(ctx, alert); err != nil {
				return fmt.Errorf("failed to create alert: %w", err)
			}

			s.logger.Info("Alert created",
				"rule_id", rule.ID,
				"rule_name", rule.Name,
				"severity", rule.Severity,
				"value", result,
				"threshold", rule.Threshold)
		}
	} else {
		// If the condition is no longer met, resolve any active alerts for this rule
		activeAlerts, err := s.alertRepo.GetAlerts(ctx, &models.AlertFilter{
			RuleID: &rule.ID,
			Status: func() *string { s := "active"; return &s }(),
		})
		if err != nil {
			return fmt.Errorf("failed to check existing alerts: %w", err)
		}

		for _, alert := range activeAlerts {
			if err := s.alertRepo.ResolveAlert(ctx, alert.ID); err != nil {
				s.logger.Error("Failed to resolve alert", "error", err, "alert_id", alert.ID)
			} else {
				s.logger.Info("Alert resolved", "alert_id", alert.ID, "rule_name", rule.Name)
			}
		}
	}

	return nil
}

// buildQuery builds the SQL query for evaluating an alert rule
func (s *AlertService) buildQuery(rule *models.AlertRule) string {
	// Add time window filter to the condition
	timeWindow := time.Now().Add(-time.Duration(rule.TimeWindow) * time.Minute)

	// Build the query with time window filter
	query := fmt.Sprintf(`
		SELECT %s 
		FROM logs 
		WHERE created_at >= '%s'
	`, rule.Condition, timeWindow.Format("2006-01-02 15:04:05"))

	return query
}

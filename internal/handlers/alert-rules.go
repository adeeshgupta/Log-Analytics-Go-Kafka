package handlers

import (
	"github.com/adeesh/log-analytics/internal/database/alert_rules"
	"github.com/adeesh/log-analytics/internal/models"
	"net/http"
	"strconv"
	"time"

	"log/slog"

	"github.com/gin-gonic/gin"
)

// AlertRuleHandler handles alert rule-related HTTP requests
type AlertRuleHandler struct {
	alertRuleRepo alert_rules.AlertRuleRepository
	logger        *slog.Logger
}

// NewAlertRuleHandler creates a new alert rule handler
func NewAlertRuleHandler(alertRuleRepo alert_rules.AlertRuleRepository, logger *slog.Logger) *AlertRuleHandler {
	return &AlertRuleHandler{
		alertRuleRepo: alertRuleRepo,
		logger:        logger,
	}
}

// CreateAlertRule creates a new alert rule
func (h *AlertRuleHandler) CreateAlertRule(c *gin.Context) {
	var rule models.AlertRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		h.logger.Error("Failed to bind alert rule", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()

	if err := h.alertRuleRepo.CreateAlertRule(c.Request.Context(), &rule); err != nil {
		h.logger.Error("Failed to create alert rule", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create alert rule"})
		return
	}

	c.JSON(http.StatusCreated, rule)
}

// GetAlertRules retrieves all alert rules
func (h *AlertRuleHandler) GetAlertRules(c *gin.Context) {
	rules, err := h.alertRuleRepo.GetAlertRules(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get alert rules", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get alert rules"})
		return
	}

	c.JSON(http.StatusOK, rules)
}

// GetAlertRuleByID retrieves an alert rule by ID
func (h *AlertRuleHandler) GetAlertRuleByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid alert rule ID"})
		return
	}

	rule, err := h.alertRuleRepo.GetAlertRuleByID(c.Request.Context(), uint(id))
	if err != nil {
		h.logger.Error("Failed to get alert rule", "error", err, "id", id)
		c.JSON(http.StatusNotFound, gin.H{"error": "Alert rule not found"})
		return
	}

	c.JSON(http.StatusOK, rule)
}

// UpdateAlertRule updates an alert rule
func (h *AlertRuleHandler) UpdateAlertRule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid alert rule ID"})
		return
	}

	var rule models.AlertRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		h.logger.Error("Failed to bind alert rule", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	rule.ID = uint(id)
	rule.UpdatedAt = time.Now()

	if err := h.alertRuleRepo.UpdateAlertRule(c.Request.Context(), &rule); err != nil {
		h.logger.Error("Failed to update alert rule", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update alert rule"})
		return
	}

	c.JSON(http.StatusOK, rule)
}

// DeleteAlertRule deletes an alert rule
func (h *AlertRuleHandler) DeleteAlertRule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid alert rule ID"})
		return
	}

	if err := h.alertRuleRepo.DeleteAlertRule(c.Request.Context(), uint(id)); err != nil {
		h.logger.Error("Failed to delete alert rule", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete alert rule"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Alert rule deleted successfully"})
}

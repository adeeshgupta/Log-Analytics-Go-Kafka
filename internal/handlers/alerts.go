package handlers

import (
	"github.com/adeesh/log-analytics/internal/database/alerts"
	"github.com/adeesh/log-analytics/internal/models"
	"net/http"
	"strconv"

	"log/slog"

	"github.com/gin-gonic/gin"
)

// AlertHandler handles alert-related HTTP requests
type AlertHandler struct {
	alertRepo alerts.AlertRepository
	logger    *slog.Logger
}

// NewAlertHandler creates a new alert handler
func NewAlertHandler(alertRepo alerts.AlertRepository, logger *slog.Logger) *AlertHandler {
	return &AlertHandler{
		alertRepo: alertRepo,
		logger:    logger,
	}
}

// GetAlerts retrieves alerts with filters
func (h *AlertHandler) GetAlerts(c *gin.Context) {
	var filter models.AlertFilter

	// Parse query parameters
	if status := c.Query("status"); status != "" {
		filter.Status = &status
	}
	if severity := c.Query("severity"); severity != "" {
		filter.Severity = &severity
	}
	if ruleIDStr := c.Query("rule_id"); ruleIDStr != "" {
		if ruleID, err := strconv.ParseUint(ruleIDStr, 10, 32); err == nil {
			ruleIDUint := uint(ruleID)
			filter.RuleID = &ruleIDUint
		}
	}
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filter.Limit = &limit
		}
	}
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			filter.Offset = &offset
		}
	}

	alerts, err := h.alertRepo.GetAlerts(c.Request.Context(), &filter)
	if err != nil {
		h.logger.Error("Failed to get alerts", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get alerts"})
		return
	}

	c.JSON(http.StatusOK, alerts)
}

// GetAlertByID retrieves an alert by ID
func (h *AlertHandler) GetAlertByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid alert ID"})
		return
	}

	alert, err := h.alertRepo.GetAlertByID(c.Request.Context(), uint(id))
	if err != nil {
		h.logger.Error("Failed to get alert", "error", err, "id", id)
		c.JSON(http.StatusNotFound, gin.H{"error": "Alert not found"})
		return
	}

	c.JSON(http.StatusOK, alert)
}

// GetAlertStats retrieves alert statistics
func (h *AlertHandler) GetAlertStats(c *gin.Context) {
	stats, err := h.alertRepo.GetAlertStats(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get alert stats", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get alert stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetActiveAlerts retrieves all active alerts
func (h *AlertHandler) GetActiveAlerts(c *gin.Context) {
	alerts, err := h.alertRepo.GetActiveAlerts(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get active alerts", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get active alerts"})
		return
	}

	c.JSON(http.StatusOK, alerts)
}

// ResolveAlert resolves an alert
func (h *AlertHandler) ResolveAlert(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid alert ID"})
		return
	}

	if err := h.alertRepo.ResolveAlert(c.Request.Context(), uint(id)); err != nil {
		h.logger.Error("Failed to resolve alert", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resolve alert"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Alert resolved successfully"})
}

// AcknowledgeAlert acknowledges an alert
func (h *AlertHandler) AcknowledgeAlert(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid alert ID"})
		return
	}

	if err := h.alertRepo.AcknowledgeAlert(c.Request.Context(), uint(id)); err != nil {
		h.logger.Error("Failed to acknowledge alert", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to acknowledge alert"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Alert acknowledged successfully"})
}

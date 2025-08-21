package handlers

import (
	"context"
	"github.com/adeesh/log-analytics/internal/database"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"log/slog"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	db     *database.GormDB
	logger *slog.Logger
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(db *database.GormDB, logger *slog.Logger) *HealthHandler {
	return &HealthHandler{
		db:     db,
		logger: logger,
	}
}

// HealthCheck performs a health check on the system
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Check database connectivity
	if err := h.db.Ping(ctx); err != nil {
		h.logger.Error("Health check failed - database ping failed", "error", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "unhealthy",
			"message": "Database connection failed",
			"timestamp": time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"message":   "Service is running",
		"timestamp": time.Now(),
		"services": gin.H{
			"database": "healthy",
			"api":      "healthy",
		},
	})
} 

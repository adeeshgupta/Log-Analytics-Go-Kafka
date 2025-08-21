package main

import (
	"context"
	"github.com/adeesh/log-analytics/internal/config"
	"github.com/adeesh/log-analytics/internal/constants"
	"github.com/adeesh/log-analytics/internal/database"
	"github.com/adeesh/log-analytics/internal/database/alert_rules"
	"github.com/adeesh/log-analytics/internal/database/alerts"
	"github.com/adeesh/log-analytics/internal/database/logs"
	"github.com/adeesh/log-analytics/internal/handlers"
	"github.com/adeesh/log-analytics/internal/services"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := database.NewGormDB(&cfg.Database)
	if err != nil {
		logger.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Create repositories
	logRepo := logs.NewLogRepository(db)
	alertRepo := alerts.NewAlertRepository(db.GetDB())
	alertRuleRepo := alert_rules.NewAlertRuleRepository(db.GetDB())

	// Create handlers
	logHandler := handlers.NewLogHandler(logRepo, logger)
	alertHandler := handlers.NewAlertHandler(alertRepo, logger)
	alertRuleHandler := handlers.NewAlertRuleHandler(alertRuleRepo, logger)
	healthHandler := handlers.NewHealthHandler(db, logger)

	// Create alert service
	sqlDB, err := db.GetSQLDB()
	if err != nil {
		logger.Error("Failed to get SQL DB", "error", err)
		os.Exit(1)
	}
	alertService := services.NewAlertService(alertRuleRepo, alertRepo, sqlDB, logger)

	// Start alert checker in background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go alertService.StartAlertChecker(ctx, time.Duration(constants.DefaultAlertCheckInterval)*time.Second)

	// Setup Gin router
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Health check endpoint
	router.GET(constants.APIHealthPath, healthHandler.HealthCheck)

	// Serve dashboard
	router.GET("/", func(c *gin.Context) {
		c.File("templates/dashboard.html")
	})

	// API routes
	api := router.Group(constants.APIPrefix)
	{
		// Log endpoints
		logsGroup := api.Group(constants.APILogsPath)
		{
			logsGroup.GET("", logHandler.GetLogs)
			logsGroup.GET("/trace/:traceID", logHandler.GetLogsByTraceID)
		}

		// Metrics endpoint for combined summary of logs
		metrics := api.Group(constants.APIMetricsPath)
		{
			metrics.GET("", logHandler.GetMetrics)
		}

		// Alert endpoints
		alertsGroup := api.Group("/alerts")
		{
			alertsGroup.GET("", alertHandler.GetAlerts)
			alertsGroup.GET("/stats", alertHandler.GetAlertStats)
			alertsGroup.GET("/active", alertHandler.GetActiveAlerts)
			alertsGroup.GET("/:id", alertHandler.GetAlertByID)
			alertsGroup.PUT("/:id/resolve", alertHandler.ResolveAlert)
			alertsGroup.PUT("/:id/acknowledge", alertHandler.AcknowledgeAlert)
		}

		// Alert rule endpoints
		rulesGroup := api.Group("/alert-rules")
		{
			rulesGroup.POST("", alertRuleHandler.CreateAlertRule)
			rulesGroup.GET("", alertRuleHandler.GetAlertRules)
			rulesGroup.GET("/:id", alertRuleHandler.GetAlertRuleByID)
			rulesGroup.PUT("/:id", alertRuleHandler.UpdateAlertRule)
			rulesGroup.DELETE("/:id", alertRuleHandler.DeleteAlertRule)
		}
	}

	//Serve static files for dashboard
	router.Static("/static", "./static")
	router.LoadHTMLGlob("templates/*")
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "dashboard.html", gin.H{
			"title": "Log Analytics Dashboard",
		})
	})

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting API server", "port", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Failed to start server", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Cancel alert checker context
	cancel()

	// Create a deadline for server shutdown
	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	}

	logger.Info("Server exited")
}

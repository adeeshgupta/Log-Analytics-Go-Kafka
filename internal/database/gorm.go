package database

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/adeesh/log-analytics/internal/config"
	"github.com/adeesh/log-analytics/internal/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// GormDB represents a GORM database connection
type GormDB struct {
	db *gorm.DB
}

// NewGormDB creates a new GORM database connection
func NewGormDB(cfg *config.DatabaseConfig) (*GormDB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// Auto migrate tables
	if err := db.AutoMigrate(
		&models.Log{},
		&models.AlertRule{},
		&models.Alert{},
	); err != nil {
		return nil, fmt.Errorf("failed to auto migrate: %w", err)
	}

	return &GormDB{db: db}, nil
}

// Close closes the database connection
func (g *GormDB) Close() error {
	sqlDB, err := g.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Ping checks if the database is accessible
func (g *GormDB) Ping(ctx context.Context) error {
	sqlDB, err := g.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

// GetDB returns the underlying GORM database instance
func (g *GormDB) GetDB() *gorm.DB {
	return g.db
}

// GetSQLDB returns the underlying sql.DB instance
func (g *GormDB) GetSQLDB() (*sql.DB, error) {
	return g.db.DB()
}

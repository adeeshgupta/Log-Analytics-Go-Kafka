package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/adeesh/log-analytics/internal/config"

	_ "github.com/go-sql-driver/mysql"
)

// Migration represents a database migration
type Migration struct {
	ID       string
	Filename string
	Content  string
}

// MigrationRunner handles database migrations
type MigrationRunner struct {
	db     *sql.DB
	logger *slog.Logger
	config *config.Config
}

// NewMigrationRunner creates a new migration runner
func NewMigrationRunner(cfg *config.Config, logger *slog.Logger) (*MigrationRunner, error) {
	// First, try to connect to MySQL server without specifying a database
	dsnWithoutDB := fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
	)

	// Connect to MySQL server
	db, err := sql.Open("mysql", dsnWithoutDB)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MySQL server: %w", err)
	}

	// Test connection to MySQL server
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping MySQL server: %w", err)
	}

	logger.Info("Connected to MySQL server successfully")

	// Configure connection pool
	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	return &MigrationRunner{
		db:     db,
		logger: logger,
		config: cfg,
	}, nil
}

// Close closes the database connection
func (m *MigrationRunner) Close() error {
	return m.db.Close()
}

// LoadMigrations loads all migration files from the migrations directory
func (m *MigrationRunner) LoadMigrations(migrationsDir string) ([]Migration, error) {
	files, err := ioutil.ReadDir(migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	var migrations []Migration
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}

		// Extract migration ID from filename (e.g., "001_initial_schema.sql" -> "001")
		parts := strings.Split(file.Name(), "_")
		if len(parts) < 2 {
			m.logger.Warn("Skipping migration file with invalid name", "filename", file.Name())
			continue
		}

		migrationID := parts[0]

		// Read migration content
		content, err := ioutil.ReadFile(filepath.Join(migrationsDir, file.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed to read migration file %s: %w", file.Name(), err)
		}

		migrations = append(migrations, Migration{
			ID:       migrationID,
			Filename: file.Name(),
			Content:  string(content),
		})
	}

	// Sort migrations by ID
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].ID < migrations[j].ID
	})

	return migrations, nil
}

// GetAppliedMigrations gets the list of already applied migrations
func (m *MigrationRunner) GetAppliedMigrations(ctx context.Context) (map[string]bool, error) {
	query := `SELECT id FROM migrations`
	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		applied[id] = true
	}

	return applied, rows.Err()
}

// ApplyMigration applies a single migration
func (m *MigrationRunner) ApplyMigration(ctx context.Context, migration Migration) error {
	m.logger.Info("Applying migration", "id", migration.ID, "filename", migration.Filename)

	// Start transaction
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Split SQL content into individual statements
	statements := m.splitSQLStatements(migration.Content)

	// Execute each statement
	for i, statement := range statements {
		statement = strings.TrimSpace(statement)
		if statement == "" {
			continue
		}

		m.logger.Debug("Executing SQL statement", "migration", migration.ID, "statement", i+1)
		if err := m.executeStatement(ctx, tx, statement); err != nil {
			return fmt.Errorf("failed to execute migration %s statement %d: %w", migration.ID, i+1, err)
		}
	}

	// For the first migration (database creation), don't record it in migrations table
	if migration.ID == "000" {
		m.logger.Info("Database creation migration completed, skipping migration record")
		// Commit transaction without recording migration
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", migration.ID, err)
		}
		return nil
	}

	// For the migrations table creation migration, don't record it either
	if migration.ID == "001" {
		m.logger.Info("Migrations table creation migration completed, skipping migration record")
		// Commit transaction without recording migration
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", migration.ID, err)
		}
		return nil
	}

	// Record migration as applied
	recordQuery := `INSERT INTO migrations (id, filename, checksum) VALUES (?, ?, ?)`
	checksum := m.generateChecksum(migration.Content)
	if _, err := tx.ExecContext(ctx, recordQuery, migration.ID, migration.Filename, checksum); err != nil {
		return fmt.Errorf("failed to record migration %s: %w", migration.ID, err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration %s: %w", migration.ID, err)
	}

	m.logger.Info("Migration applied successfully", "id", migration.ID, "filename", migration.Filename)
	return nil
}

// splitSQLStatements splits SQL content into individual statements
func (m *MigrationRunner) splitSQLStatements(content string) []string {
	// Remove comments
	lines := strings.Split(content, "\n")
	var cleanLines []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "--") {
			continue
		}
		cleanLines = append(cleanLines, line)
	}

	// Join lines and split by semicolon
	cleanContent := strings.Join(cleanLines, " ")
	statements := strings.Split(cleanContent, ";")

	var result []string
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt != "" {
			result = append(result, stmt)
		}
	}

	return result
}

// executeStatement executes a single SQL statement, handling USE statements specially
func (m *MigrationRunner) executeStatement(ctx context.Context, tx *sql.Tx, statement string) error {
	statement = strings.TrimSpace(statement)
	if statement == "" {
		return nil
	}

	// Handle USE statements by executing them directly on the connection
	if strings.HasPrefix(strings.ToUpper(statement), "USE ") {
		// Execute USE statement on the connection, not in transaction
		if _, err := m.db.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("failed to execute USE statement: %w", err)
		}
		return nil
	}

	// Execute other statements in the transaction
	if _, err := tx.ExecContext(ctx, statement); err != nil {
		return fmt.Errorf("failed to execute statement: %w", err)
	}

	return nil
}

// generateChecksum generates a SHA256 hash of the content
func (m *MigrationRunner) generateChecksum(content string) string {
	hash := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", hash)
}

// RunMigrations runs all pending migrations
func (m *MigrationRunner) RunMigrations(migrationsDir string) error {
	ctx := context.Background()

	// Load migrations
	migrations, err := m.LoadMigrations(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	if len(migrations) == 0 {
		m.logger.Info("No migration files found")
		return nil
	}

	m.logger.Info("Found migrations", "count", len(migrations))
	for _, migration := range migrations {
		m.logger.Debug("Migration file", "id", migration.ID, "filename", migration.Filename)
	}

	// Get applied migrations (only after migrations table exists)
	applied := make(map[string]bool)
	if len(migrations) > 0 && (migrations[0].ID == "000" || migrations[0].ID == "001") {
		// First migrations create database and migrations table, so we can't check applied migrations yet
		m.logger.Info("First migrations will create database and migrations table, skipping applied migrations check")
	} else {
		// Get applied migrations
		applied, err = m.GetAppliedMigrations(ctx)
		if err != nil {
			return fmt.Errorf("failed to get applied migrations: %w", err)
		}
	}

	// Apply pending migrations
	appliedCount := 0
	for i, migration := range migrations {
		// For the first two migrations (database creation and migrations table), always apply them
		if i <= 1 && (migration.ID == "000" || migration.ID == "001") {
			if err := m.ApplyMigration(ctx, migration); err != nil {
				return fmt.Errorf("failed to apply migration %s: %w", migration.ID, err)
			}
			appliedCount++

			// After first migration, reconnect to the specific database
			if i == 0 && migration.ID == "000" {
				m.logger.Info("Reconnecting to specific database after creation")
				if err := m.reconnectToDatabase(); err != nil {
					return fmt.Errorf("failed to reconnect to database: %w", err)
				}
			}
			continue
		}

		// Before handling subsequent migrations, ensure we have the applied set
		// This covers the case where we just created the DB and migrations table
		if len(applied) == 0 {
			var err error
			applied, err = m.GetAppliedMigrations(ctx)
			if err != nil {
				return fmt.Errorf("failed to get applied migrations: %w", err)
			}
		}

		// For subsequent migrations, check if already applied
		if applied[migration.ID] {
			m.logger.Debug("Migration already applied", "id", migration.ID, "filename", migration.Filename)
			continue
		}

		if err := m.ApplyMigration(ctx, migration); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", migration.ID, err)
		}

		appliedCount++
	}

	m.logger.Info("Migrations completed", "applied", appliedCount, "total", len(migrations))
	return nil
}

// reconnectToDatabase reconnects to the specific database after it's created
func (m *MigrationRunner) reconnectToDatabase() error {
	// Close current connection
	m.db.Close()

	// Connect to the specific database
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		m.config.Database.Username,
		m.config.Database.Password,
		m.config.Database.Host,
		m.config.Database.Port,
		m.config.Database.Database,
	)

	// Connect to the specific database
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test connection to the specific database
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(m.config.Database.MaxOpenConns)
	db.SetMaxIdleConns(m.config.Database.MaxIdleConns)
	db.SetConnMaxLifetime(m.config.Database.ConnMaxLifetime)

	// Update the connection
	m.db = db

	m.logger.Info("Reconnected to database successfully", "database", m.config.Database.Database)
	return nil
}

// ShowStatus shows the current migration status
func (m *MigrationRunner) ShowStatus(migrationsDir string) error {
	ctx := context.Background()

	// Load migrations
	migrations, err := m.LoadMigrations(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	// Try to connect to the specific database first
	originalDB := m.db
	if err := m.reconnectToDatabase(); err != nil {
		// Database doesn't exist yet, all migrations are pending
		m.logger.Info("Database does not exist yet, all migrations are pending")
		for _, migration := range migrations {
			m.logger.Info("Migration", "id", migration.ID, "filename", migration.Filename, "status", "PENDING")
		}
		// Restore original connection
		m.db = originalDB
		return nil
	}

	// Get applied migrations
	applied, err := m.GetAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	m.logger.Info("Migration Status", "total", len(migrations))
	for _, migration := range migrations {
		status := "PENDING"
		if applied[migration.ID] {
			status = "APPLIED"
		}
		m.logger.Info("Migration", "id", migration.ID, "filename", migration.Filename, "status", status)
	}

	return nil
}

// SetupDatabase performs complete database setup
func (m *MigrationRunner) SetupDatabase(migrationsDir string) error {
	m.logger.Info("Starting complete database setup")

	// Run migrations (this will create the database and migrations table)
	if err := m.RunMigrations(migrationsDir); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	m.logger.Info("Database setup completed successfully")
	return nil
}

func main() {
	// Initialize logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Load configuration
	cfg := config.Load()

	// Get command line arguments
	args := os.Args[1:]
	if len(args) == 0 {
		args = []string{"run"}
	}

	command := args[0]
	migrationsDir := "scripts/migrations"

	switch command {
	case "setup":
		logger.Info("Setting up database and running migrations")
		// Create migration runner
		runner, err := NewMigrationRunner(cfg, logger)
		if err != nil {
			logger.Error("Failed to create migration runner", "error", err)
			os.Exit(1)
		}
		defer runner.Close()

		if err := runner.SetupDatabase(migrationsDir); err != nil {
			logger.Error("Failed to setup database", "error", err)
			os.Exit(1)
		}
		logger.Info("Database setup completed successfully")

	case "run":
		logger.Info("Running database migrations")
		// Create migration runner
		runner, err := NewMigrationRunner(cfg, logger)
		if err != nil {
			logger.Error("Failed to create migration runner", "error", err)
			os.Exit(1)
		}
		defer runner.Close()

		if err := runner.RunMigrations(migrationsDir); err != nil {
			logger.Error("Failed to run migrations", "error", err)
			os.Exit(1)
		}
		logger.Info("Migrations completed successfully")

	case "status":
		logger.Info("Showing migration status")
		// Create migration runner
		runner, err := NewMigrationRunner(cfg, logger)
		if err != nil {
			logger.Error("Failed to create migration runner", "error", err)
			os.Exit(1)
		}
		defer runner.Close()

		if err := runner.ShowStatus(migrationsDir); err != nil {
			logger.Error("Failed to show status", "error", err)
			os.Exit(1)
		}

	default:
		logger.Error("Unknown command", "command", command)
		logger.Info("Available commands: setup, run, status")
		logger.Info("  setup  - Complete database setup (creates DB and runs migrations)")
		logger.Info("  run    - Run pending migrations only")
		logger.Info("  status - Show migration status")
		os.Exit(1)
	}
}

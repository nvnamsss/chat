package adapters

import (
	"context"
	"fmt"

	"github.com/nvnamsss/chat/src/configs"
	"github.com/nvnamsss/chat/src/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DBAdapter defines the interface for database operations
type DBAdapter interface {
	GetDB() *gorm.DB
	Close() error
	Ping(ctx context.Context) error
	AutoMigrate(models ...interface{}) error
}

// dbAdapter implements the DBAdapter interface
type dbAdapter struct {
	db *gorm.DB
}

// NewDBAdapter creates a new database adapter
func NewDBAdapter(config configs.Database) (DBAdapter, error) {
	// Configure GORM
	gormConfig := &gorm.Config{
		Logger: logger.NewGormLogger(),
	}

	// Connect to database
	db, err := gorm.Open(postgres.Open(config.DSN()), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying *sql.DB: %w", err)
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)

	logger.Info("Connected to database",
		logger.Field("host", config.Host),
		logger.Field("database", config.Name))

	return &dbAdapter{db: db}, nil
}

// GetDB returns the database connection
func (a *dbAdapter) GetDB() *gorm.DB {
	return a.db
}

// Close closes the database connection
func (a *dbAdapter) Close() error {
	logger.Info("Closing database connection")
	sqlDB, err := a.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying *sql.DB: %w", err)
	}
	return sqlDB.Close()
}

// Ping checks the database connection
func (a *dbAdapter) Ping(ctx context.Context) error {
	sqlDB, err := a.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying *sql.DB: %w", err)
	}
	return sqlDB.PingContext(ctx)
}

// AutoMigrate runs GORM auto-migration for the given models
func (a *dbAdapter) AutoMigrate(models ...interface{}) error {
	return a.db.AutoMigrate(models...)
}

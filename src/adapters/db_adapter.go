package adapters

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/nvnamsss/chat/src/configs"
	"github.com/nvnamsss/chat/src/logger"
)

// DBAdapter defines the interface for database operations
type DBAdapter interface {
	GetDB() *sqlx.DB
	Close() error
	Ping(ctx context.Context) error
}

// dbAdapter implements the DBAdapter interface
type dbAdapter struct {
	db *sqlx.DB
}

// NewDBAdapter creates a new database adapter
func NewDBAdapter(config configs.Database) (DBAdapter, error) {
	db, err := sqlx.Connect("postgres", config.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	logger.Info("Connected to database",
		logger.Field("host", config.Host),
		logger.Field("database", config.Name))

	return &dbAdapter{db: db}, nil
}

// GetDB returns the database connection
func (a *dbAdapter) GetDB() *sqlx.DB {
	return a.db
}

// Close closes the database connection
func (a *dbAdapter) Close() error {
	logger.Info("Closing database connection")
	return a.db.Close()
}

// Ping checks the database connection
func (a *dbAdapter) Ping(ctx context.Context) error {
	return a.db.PingContext(ctx)
}

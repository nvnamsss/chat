package logger

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// GormLogger implements GORM's logger.Interface
type GormLogger struct {
	SlowThreshold         time.Duration
	SourceField           string
	SkipCallerLookup      bool
	SkipErrRecordNotFound bool
}

// NewGormLogger creates a new GORM logger that integrates with our logging system
func NewGormLogger() logger.Interface {
	return &GormLogger{
		SlowThreshold:         200 * time.Millisecond,
		SkipErrRecordNotFound: true,
		SourceField:           "source",
	}
}

// LogMode sets the logger mode
func (l *GormLogger) LogMode(level logger.LogLevel) logger.Interface {
	return l
}

// Info logs info messages
func (l *GormLogger) Info(ctx context.Context, msg string, args ...interface{}) {
	Context(ctx).Infof(msg, args...)
}

// Warn logs warning messages
func (l *GormLogger) Warn(ctx context.Context, msg string, args ...interface{}) {
	Context(ctx).Warnf(msg, args...)
}

// Error logs error messages
func (l *GormLogger) Error(ctx context.Context, msg string, args ...interface{}) {
	Context(ctx).Errorf(msg, args...)
}

// Trace logs SQL queries
func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()

	// Skip logging if no error and query is fast
	if err == nil && elapsed < l.SlowThreshold {
		return
	}

	// Skip "record not found" errors if configured
	if l.SkipErrRecordNotFound && errors.Is(err, gorm.ErrRecordNotFound) {
		return
	}

	// Log slow queries
	if elapsed > l.SlowThreshold {
		Context(ctx).Warnw("Slow SQL query",
			"elapsed", elapsed,
			"sql", sql,
			"rows", rows,
			"error", err,
		)
		return
	}

	// Log queries with errors
	if err != nil {
		Context(ctx).Errorw("Failed SQL query",
			"elapsed", elapsed,
			"sql", sql,
			"rows", rows,
			"error", err,
		)
		return
	}

	// Log normal queries at debug level
	Context(ctx).Debugw("SQL query",
		"elapsed", elapsed,
		"sql", sql,
		"rows", rows,
	)
}

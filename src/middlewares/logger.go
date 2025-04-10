package middlewares

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nvnamsss/chat/src/logger"
)

// Logger returns a middleware that logs HTTP requests
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get request ID from context
		ctx := logger.WithRequestID(c.Request.Context())
		c.Request = c.Request.WithContext(ctx)

		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get request ID for logging
		reqID := logger.GetRequestID(ctx)

		// Log request details
		log := logger.Context(ctx)
		log.Infow("HTTP Request",
			"status", c.Writer.Status(),
			"method", c.Request.Method,
			"path", path,
			"query", raw,
			"ip", c.ClientIP(),
			"latency", latency,
			"user-agent", c.Request.UserAgent(),
			"request_id", reqID,
			"errors", c.Errors.String(),
		)
	}
}

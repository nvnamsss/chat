package middlewares

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nvnamsss/chat/src/logger"
)

// RequestID returns a middleware that adds a unique request ID to each request
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if request ID already exists in headers
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			// Generate a new request ID
			requestID = uuid.New().String()
		}

		// Set request ID in context and headers
		c.Set("RequestID", requestID)
		c.Header("X-Request-ID", requestID)

		// Add request ID to the context for logging
		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, logger.RequestIDKey, requestID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

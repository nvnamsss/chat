package middlewares

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/nvnamsss/chat/src/errors"
	"github.com/nvnamsss/chat/src/logger"
)

// Auth returns a middleware for JWT authentication
func Auth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := logger.Context(c.Request.Context())

		// Skip auth for health check
		if c.Request.URL.Path == "/health" {
			c.Next()
			return
		}

		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Warnw("Missing Authorization header")
			c.AbortWithStatusJSON(401, gin.H{
				"code":    errors.ErrUnauthorized,
				"message": "Missing authentication token",
			})
			return
		}

		// Check if the header has the expected format
		authParts := strings.Split(authHeader, " ")
		if len(authParts) != 2 || authParts[0] != "Bearer" {
			log.Warnw("Invalid Authorization header format")
			c.AbortWithStatusJSON(401, gin.H{
				"code":    errors.ErrUnauthorized,
				"message": "Invalid authentication token format",
			})
			return
		}

		// Parse and validate token
		tokenStr := authParts[1]
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			// Validate the signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			log.Warnw("Invalid authentication token", "error", err)
			c.AbortWithStatusJSON(401, gin.H{
				"code":    errors.ErrUnauthorized,
				"message": "Invalid or expired authentication token",
			})
			return
		}

		// Extract claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			log.Warnw("Failed to extract token claims")
			c.AbortWithStatusJSON(401, gin.H{
				"code":    errors.ErrUnauthorized,
				"message": "Invalid token claims",
			})
			return
		}

		// Extract user ID from claims
		userID, ok := claims["sub"].(string)
		if !ok {
			log.Warnw("Missing user ID in token")
			c.AbortWithStatusJSON(401, gin.H{
				"code":    errors.ErrUnauthorized,
				"message": "Invalid user identification",
			})
			return
		}

		// Store user ID in context
		c.Set("userID", userID)

		// Store claims in context if needed
		c.Set("claims", claims)

		c.Next()
	}
}

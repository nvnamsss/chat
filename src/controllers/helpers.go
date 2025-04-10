package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nvnamsss/chat/src/errors"
	"github.com/nvnamsss/chat/src/logger"
)

// ErrorResponse represents the structure of error responses
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// respondError sends an error response to the client
func respondError(c *gin.Context, err error) {
	log := logger.Context(c.Request.Context())

	var statusCode int
	var errorResponse ErrorResponse

	// Check if this is an application error
	if appErr, ok := err.(*errors.AppError); ok {
		statusCode = appErr.StatusCode()
		errorResponse = ErrorResponse{
			Code:    appErr.Code,
			Message: appErr.Message,
		}
		log.Warnw("Application error", "code", appErr.Code, "message", appErr.Message, "error", appErr.Err)
	} else {
		// Unknown error
		statusCode = http.StatusInternalServerError
		errorResponse = ErrorResponse{
			Code:    errors.ErrInternal,
			Message: "Internal server error",
		}
		log.Errorw("Unknown error", "error", err)
	}

	c.JSON(statusCode, errorResponse)
}

package errors

import (
	"fmt"
	"net/http"
)

// Error codes
const (
	ErrInvalidRequest = "INVALID_REQUEST"
	ErrNotFound       = "NOT_FOUND"
	ErrInternal       = "INTERNAL_ERROR"
	ErrUnauthorized   = "UNAUTHORIZED"
	ErrForbidden      = "FORBIDDEN"
	ErrLLMService     = "LLM_SERVICE_ERROR"
)

// AppError represents an application error
type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// StatusCode returns the HTTP status code associated with the error
func (e *AppError) StatusCode() int {
	switch e.Code {
	case ErrInvalidRequest:
		return http.StatusBadRequest
	case ErrNotFound:
		return http.StatusNotFound
	case ErrUnauthorized:
		return http.StatusUnauthorized
	case ErrForbidden:
		return http.StatusForbidden
	case ErrLLMService:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

// New creates a new AppError
func New(code string, msg ...interface{}) *AppError {
	var message string
	var err error

	if len(msg) == 0 {
		message = getDefaultMessage(code)
	} else if len(msg) == 1 {
		switch m := msg[0].(type) {
		case string:
			message = m
		case error:
			message = getDefaultMessage(code)
			err = m
		default:
			message = fmt.Sprintf("%v", m)
		}
	} else if len(msg) == 2 {
		message = fmt.Sprintf("%v", msg[0])
		if e, ok := msg[1].(error); ok {
			err = e
		}
	}

	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// Wrap wraps an existing error in an AppError
func Wrap(err error, code string, msg ...interface{}) *AppError {
	appErr := New(code, msg...)
	appErr.Err = err
	return appErr
}

// getDefaultMessage returns a default message for a given error code
func getDefaultMessage(code string) string {
	switch code {
	case ErrInvalidRequest:
		return "Invalid request parameters"
	case ErrNotFound:
		return "Resource not found"
	case ErrInternal:
		return "Internal server error"
	case ErrUnauthorized:
		return "Unauthorized access"
	case ErrForbidden:
		return "Access forbidden"
	case ErrLLMService:
		return "LLM service error"
	default:
		return "An error occurred"
	}
}

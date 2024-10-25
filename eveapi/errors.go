package eveapi

import (
	"fmt"
	"net/http"
)

// CustomError represents an HTTP error with a status code and message
type CustomError struct {
	StatusCode int
	Message    string
}

func (e *CustomError) Error() string {
	return fmt.Sprintf("%d: %s", e.StatusCode, e.Message)
}

func NewCustomError(statusCode int, message string) *CustomError {
	return &CustomError{StatusCode: statusCode, Message: message}
}

// Map of HTTP status codes to custom errors
var httpStatusErrors = map[int]*CustomError{
	http.StatusForbidden:           ErrForbidden,
	http.StatusServiceUnavailable:  ErrServiceUnavailable,
	http.StatusGatewayTimeout:      ErrGatewayTimeout,
	http.StatusInternalServerError: ErrInternalServerError,
	http.StatusBadRequest:          ErrBadRequest,
	http.StatusUnauthorized:        ErrUnauthorized,
	http.StatusNotFound:            ErrNotFound,
	http.StatusMethodNotAllowed:    ErrMethodNotAllowed,
	http.StatusBadGateway:          ErrBadGateway,
}

// Predefined errors for common HTTP status codes
var (
	ErrBadRequest          = NewCustomError(http.StatusBadRequest, "bad request")
	ErrUnauthorized        = NewCustomError(http.StatusUnauthorized, "unauthorized")
	ErrForbidden           = NewCustomError(http.StatusForbidden, "forbidden")
	ErrNotFound            = NewCustomError(http.StatusNotFound, "not found")
	ErrMethodNotAllowed    = NewCustomError(http.StatusMethodNotAllowed, "method not allowed")
	ErrInternalServerError = NewCustomError(http.StatusInternalServerError, "internal server error")
	ErrBadGateway          = NewCustomError(http.StatusBadGateway, "bad gateway")
	ErrServiceUnavailable  = NewCustomError(http.StatusServiceUnavailable, "service unavailable")
	ErrGatewayTimeout      = NewCustomError(http.StatusGatewayTimeout, "gateway timeout")
)

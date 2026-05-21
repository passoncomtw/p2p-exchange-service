package errors

import "net/http"

type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *AppError) Error() string {
	return e.Message
}

func New(code int, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

var (
	ErrUnauthorized       = New(http.StatusUnauthorized, "unauthorized")
	ErrInvalidCredentials = New(http.StatusUnauthorized, "invalid username or password")
	ErrForbidden          = New(http.StatusForbidden, "forbidden")
	ErrNotFound           = New(http.StatusNotFound, "not found")
	ErrInternal           = New(http.StatusInternalServerError, "internal server error")
)

package application

import (
	"errors"
	"net/http"
)

type AppError struct {
	Err        error
	HTTPStatus int
	Code       string
}

func (e *AppError) Error() string { return e.Err.Error() }
func (e *AppError) Unwrap() error { return e.Err }

var (
	ErrNotFound     = &AppError{Err: errors.New("resource not found"), HTTPStatus: http.StatusNotFound, Code: "NOT_FOUND"}
	ErrValidation   = &AppError{Err: errors.New("validation failed"), HTTPStatus: http.StatusBadRequest, Code: "VALIDATION_ERROR"}
	ErrConflict     = &AppError{Err: errors.New("resource conflict"), HTTPStatus: http.StatusConflict, Code: "CONFLICT"}
	ErrRateLimited  = &AppError{Err: errors.New("rate limit exceeded"), HTTPStatus: http.StatusTooManyRequests, Code: "RATE_LIMITED"}
	ErrUnauthorized = &AppError{Err: errors.New("unauthorized"), HTTPStatus: http.StatusUnauthorized, Code: "UNAUTHORIZED"}
	ErrForbidden    = &AppError{Err: errors.New("forbidden"), HTTPStatus: http.StatusForbidden, Code: "FORBIDDEN"}
	ErrTimeout      = &AppError{Err: errors.New("operation timed out"), HTTPStatus: http.StatusGatewayTimeout, Code: "TIMEOUT"}
)

package httpapi

import (
	"errors"

	"github.com/flyluman/scratch/internal/application"
	"github.com/flyluman/scratch/internal/ports"
	"github.com/labstack/echo/v4"
)

type Envelope struct {
	Success bool     `json:"success"`
	Data    any      `json:"data,omitempty"`
	Error   *ErrBody `json:"error,omitempty"`
}

type ErrBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func WriteSuccess(c echo.Context, status int, data any) error {
	return c.JSON(status, Envelope{Success: true, Data: data})
}

func WriteAppError(c echo.Context, err error) error {
	var appErr *application.AppError
	if errors.As(err, &appErr) {
		return c.JSON(appErr.HTTPStatus, Envelope{Success: false, Error: &ErrBody{Code: appErr.Code, Message: appErr.Error()}})
	}
	return c.JSON(500, Envelope{Success: false, Error: &ErrBody{Code: "INTERNAL_ERROR", Message: err.Error()}})
}

func ActorID(c echo.Context) string {
	claims, ok := c.Get("claims").(ports.Claims)
	if !ok {
		return ""
	}
	return claims.Subject
}

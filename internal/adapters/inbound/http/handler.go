package httpapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/flyluman/scratch/internal/platform/config"
	"github.com/flyluman/scratch/internal/platform/telemetry"
	"github.com/flyluman/scratch/internal/ports"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/swaggo/swag"
	"golang.org/x/time/rate"
)

type HealthCheck func(ctx context.Context) error

type Handler struct {
	subHandlers    []RouteRegistrar
	cfg            config.Config
	logger         ports.Logger
	tel            *telemetry.Telemetry
	tokenValidator ports.TokenValidator
	healthChecks   []HealthCheck
}

type Option func(*Handler)

func WithLogger(l ports.Logger) Option {
	return func(h *Handler) { h.logger = l }
}

func WithTelemetry(t *telemetry.Telemetry) Option {
	return func(h *Handler) { h.tel = t }
}

func WithTokenValidator(v ports.TokenValidator) Option {
	return func(h *Handler) { h.tokenValidator = v }
}

func WithHealthChecks(hc ...HealthCheck) Option {
	return func(h *Handler) { h.healthChecks = hc }
}

func NewHandler(cfg config.Config, subHandlers []RouteRegistrar, opts ...Option) *Handler {
	h := &Handler{
		subHandlers: subHandlers,
		cfg:         cfg,
	}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

func (h *Handler) RegisterRoutes(e *echo.Echo) {
	e.HTTPErrorHandler = h.errorHandler

	e.Use(
		middleware.Recover(),
		middleware.RequestID(),
		middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins: []string{"*"},
			AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
		}),
		middleware.LoggerWithConfig(middleware.LoggerConfig{}),
		middleware.TimeoutWithConfig(middleware.TimeoutConfig{Timeout: 30 * time.Second}),
		middleware.BodyLimit("1MB"),
		middleware.Secure(),
	)

	if h.cfg.RateLimitPerSecond > 0 {
		e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(
			rate.Limit(h.cfg.RateLimitPerSecond),
		)))
	}

	e.GET("/livez", h.livez)
	e.GET("/readyz", h.readyz)

	e.GET("/scalar", func(c echo.Context) error {
		return c.HTML(http.StatusOK, `<!doctype html>
<html>
<head><title>Scratch API</title></head>
<body>
<script id="api-reference" data-url="/scalar/openapi.json"></script>
<script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
</body>
</html>`)
	})
	e.GET("/scalar/openapi.json", func(c echo.Context) error {
		doc, err := swag.ReadDoc("swagger")
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		return c.String(http.StatusOK, doc)
	})

	v1 := e.Group("/v1")
	if h.tokenValidator != nil {
		v1.Use(h.authMiddleware)
	}
	for _, sh := range h.subHandlers {
		sh.RegisterRoutes(v1)
	}
}

func (h *Handler) livez(c echo.Context) error {
	return c.NoContent(http.StatusOK)
}

func (h *Handler) readyz(c echo.Context) error {
	for _, check := range h.healthChecks {
		if err := check(c.Request().Context()); err != nil {
			return c.JSON(http.StatusServiceUnavailable, Envelope{
				Success: false,
				Error:   &ErrBody{Code: "UNHEALTHY", Message: err.Error()},
			})
		}
	}
	return c.NoContent(http.StatusOK)
}

func (h *Handler) authMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return c.JSON(http.StatusUnauthorized, Envelope{
				Success: false,
				Error:   &ErrBody{Code: "UNAUTHORIZED", Message: "missing or invalid authorization header"},
			})
		}
		token := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := h.tokenValidator.Validate(c.Request().Context(), token)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, Envelope{
				Success: false,
				Error:   &ErrBody{Code: "UNAUTHORIZED", Message: err.Error()},
			})
		}
		c.Set("claims", claims)
		return next(c)
	}
}

func (h *Handler) errorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	var httpErr *echo.HTTPError
	if errors.As(err, &httpErr) {
		code := httpErr.Code
		msg := fmt.Sprintf("%v", httpErr.Message)
		_ = c.JSON(code, Envelope{Success: false, Error: &ErrBody{
			Code:    errCodeFromStatus(code),
			Message: msg,
		}})
		return
	}

	WriteAppError(c, err)
}

func errCodeFromStatus(code int) string {
	switch code {
	case 400:
		return "BAD_REQUEST"
	case 401:
		return "UNAUTHORIZED"
	case 403:
		return "FORBIDDEN"
	case 404:
		return "NOT_FOUND"
	case 405:
		return "METHOD_NOT_ALLOWED"
	case 409:
		return "CONFLICT"
	case 422:
		return "UNPROCESSABLE_ENTITY"
	case 429:
		return "TOO_MANY_REQUESTS"
	default:
		return http.StatusText(code)
	}
}

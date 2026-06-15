package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/flyluman/scratch/internal/platform/config"
	"github.com/labstack/echo/v4"
)

func executeRequest(h *Handler, method, path string, body string) *httptest.ResponseRecorder {
	e := echo.New()
	h.RegisterRoutes(e)

	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func TestLivez(t *testing.T) {
	h := NewHandler(config.Config{HTTPAddr: ":0"}, nil)
	rec := executeRequest(h, http.MethodGet, "/livez", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestReadyz(t *testing.T) {
	h := NewHandler(config.Config{HTTPAddr: ":0"}, nil)
	rec := executeRequest(h, http.MethodGet, "/readyz", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

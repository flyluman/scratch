package item

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/flyluman/scratch/internal/adapters/outbound/memory"
	"github.com/flyluman/scratch/internal/application"
	"github.com/flyluman/scratch/internal/platform/bus"
	"github.com/flyluman/scratch/internal/platform/featureenv"
	clock "github.com/flyluman/scratch/internal/provider/clock"
	idadapter "github.com/flyluman/scratch/internal/provider/id"
	"github.com/labstack/echo/v4"
)

type envelope struct {
	Success bool                   `json:"success"`
	Data    map[string]any         `json:"data,omitempty"`
	Error   map[string]any         `json:"error,omitempty"`
}

func setupHandler(t *testing.T) *Handler {
	t.Helper()

	repo := memory.NewItemRepository()
	idGen := idadapter.NewUUIDGenerator()
	clk := clock.NewRealClock()
	eventBus := bus.NewMemoryBus()
	flags := featureenv.LoadFlagsFromEnv()

	mod := application.NewItemModule(repo, idGen, clk, eventBus, flags)
	return NewHandler(mod)
}

func executeRequest(t *testing.T, h *Handler, method, path string, body string) *httptest.ResponseRecorder {
	t.Helper()

	e := echo.New()
	v1 := e.Group("/v1")
	h.RegisterRoutes(v1)

	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func readResponse(t *testing.T, rec *httptest.ResponseRecorder) envelope {
	t.Helper()

	var env envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	return env
}

func TestCreateItem(t *testing.T) {
	h := setupHandler(t)
	rec := executeRequest(t, h, http.MethodPost, "/v1/items", `{"name":"test","description":"hello"}`)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	env := readResponse(t, rec)
	if !env.Success {
		t.Fatalf("expected success=true, got response: %s", rec.Body.String())
	}
}

func TestCreateItemValidation(t *testing.T) {
	h := setupHandler(t)
	rec := executeRequest(t, h, http.MethodPost, "/v1/items", `{"name":"","description":""}`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}

	env := readResponse(t, rec)
	if env.Success {
		t.Fatal("expected success=false for validation error")
	}
	if env.Error == nil || env.Error["code"] != "VALIDATION_ERROR" {
		t.Fatalf("expected VALIDATION_ERROR, got: %+v", env.Error)
	}
}

func TestGetItem(t *testing.T) {
	h := setupHandler(t)

	createRec := executeRequest(t, h, http.MethodPost, "/v1/items", `{"name":"test","description":"hello"}`)
	env := readResponse(t, createRec)

	data := env.Data
	id := data["id"].(string)

	rec := executeRequest(t, h, http.MethodGet, "/v1/items/"+id, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestGetItemNotFound(t *testing.T) {
	h := setupHandler(t)
	rec := executeRequest(t, h, http.MethodGet, "/v1/items/nonexistent", "")

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestListItem(t *testing.T) {
	h := setupHandler(t)

	executeRequest(t, h, http.MethodPost, "/v1/items", `{"name":"test","description":"hello"}`)

	rec := executeRequest(t, h, http.MethodGet, "/v1/items", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	env := readResponse(t, rec)
	if env.Data == nil {
		t.Fatal("expected data in response")
	}
}

func TestUpdateItem(t *testing.T) {
	h := setupHandler(t)

	createRec := executeRequest(t, h, http.MethodPost, "/v1/items", `{"name":"test","description":"hello"}`)
	env := readResponse(t, createRec)
	id := env.Data["id"].(string)

	rec := executeRequest(t, h, http.MethodPut, "/v1/items/"+id, `{"name":"updated","description":"world"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDeleteItem(t *testing.T) {
	h := setupHandler(t)

	createRec := executeRequest(t, h, http.MethodPost, "/v1/items", `{"name":"test","description":"hello"}`)
	env := readResponse(t, createRec)
	id := env.Data["id"].(string)

	rec := executeRequest(t, h, http.MethodDelete, "/v1/items/"+id, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	getRec := executeRequest(t, h, http.MethodGet, "/v1/items/"+id, "")
	if getRec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 after delete, got %d", getRec.Code)
	}
}

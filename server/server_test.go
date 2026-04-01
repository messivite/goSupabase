package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/mustafaaksoy/gosupabase/handlers"
)

func TestNewHandlerRegistersAndServesRoute(t *testing.T) {
	handlers.Register("TestHealth", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})

	dir := t.TempDir()
	apiPath := filepath.Join(dir, "api.yaml")
	content := `version: "1"
basePath: /api
endpoints:
  - method: GET
    path: /health
    handler: TestHealth
    auth: false
`
	if err := os.WriteFile(apiPath, []byte(content), 0644); err != nil {
		t.Fatalf("write api.yaml: %v", err)
	}

	h := NewHandler(apiPath, "", "", "auto")
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusTeapot {
		t.Fatalf("expected 418, got %d", rr.Code)
	}
}

func TestNewHandlerMissingFileReturnsRouter(t *testing.T) {
	h := NewHandler("does-not-exist.yaml", "", "", "auto")
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code == 0 {
		t.Fatal("expected a valid HTTP status code")
	}
}

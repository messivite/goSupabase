package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandlerDoesNotPanic(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rr := httptest.NewRecorder()
	Handler(rr, req)
	if rr.Code == 0 {
		t.Fatal("expected a valid HTTP status code")
	}
}

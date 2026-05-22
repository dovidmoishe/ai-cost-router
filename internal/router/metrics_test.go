package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestResponseTimeMiddleware(t *testing.T) {
	handler := ResponseTime(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Header().Get("X-Response-Time-Ms") == "" {
		t.Fatal("expected X-Response-Time-Ms header")
	}
}

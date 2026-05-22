package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRouteHandlerAcceptsJSONWithCharset(t *testing.T) {
	Init()

	req := httptest.NewRequest(http.MethodPost, "/v1/route", strings.NewReader(`{"userText":"hey"}`))
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	rec := httptest.NewRecorder()

	RouteHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d want 200 body=%s", rec.Code, rec.Body.String())
	}
}

func TestRouteHandlerRequiresJSONContentType(t *testing.T) {
	Init()

	req := httptest.NewRequest(http.MethodPost, "/v1/route", strings.NewReader(`{"userText":"hey"}`))
	rec := httptest.NewRecorder()

	RouteHandler(rec, req)

	if rec.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("status: got %d want 415", rec.Code)
	}
}

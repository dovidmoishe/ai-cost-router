package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestInternalServiceAuth(t *testing.T) {
	t.Setenv("INTERNAL_SERVICE_TOKEN", "test-secret-token")
	loadInternalServiceToken()

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := InternalServiceAuth(next)

	t.Run("valid bearer token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/v1/route", nil)
		req.Header.Set("Authorization", "Bearer test-secret-token")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status: got %d want 200", rec.Code)
		}
	})

	t.Run("missing authorization header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/v1/route", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		assertUnauthorized(t, rec, "missing or invalid Authorization header")
	})

	t.Run("invalid bearer token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/v1/route", nil)
		req.Header.Set("Authorization", "Bearer wrong-token")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		assertUnauthorized(t, rec, "invalid authorization token")
	})

	t.Run("malformed authorization header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/v1/route", nil)
		req.Header.Set("Authorization", "Token test-secret-token")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		assertUnauthorized(t, rec, "missing or invalid Authorization header")
	})
}

func TestInternalServiceAuthNotConfigured(t *testing.T) {
	t.Setenv("INTERNAL_SERVICE_TOKEN", "")
	loadInternalServiceToken()

	handler := InternalServiceAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/v1/route", nil)
	req.Header.Set("Authorization", "Bearer anything")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assertUnauthorized(t, rec, "internal service auth is not configured")
}

func TestBearerToken(t *testing.T) {
	token, ok := bearerToken("Bearer abc123")
	if !ok || token != "abc123" {
		t.Fatalf("bearerToken: got (%q, %v)", token, ok)
	}
	if _, ok := bearerToken(""); ok {
		t.Fatal("expected empty header to fail")
	}
	if _, ok := bearerToken("Bearer "); ok {
		t.Fatal("expected empty bearer value to fail")
	}
}

func assertUnauthorized(t *testing.T, rec *httptest.ResponseRecorder, wantMessage string) {
	t.Helper()
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status: got %d want 401", rec.Code)
	}
	var body ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Error != "unauthorized" {
		t.Fatalf("error: got %q want unauthorized", body.Error)
	}
	if body.Message != wantMessage {
		t.Fatalf("message: got %q want %q", body.Message, wantMessage)
	}
}

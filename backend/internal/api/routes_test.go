package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORSMiddlewareAllowsLocalhostOrigin(t *testing.T) {
	handler := withCORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called for preflight requests")
	}))

	request := httptest.NewRequest(http.MethodOptions, "/", nil)
	request.Header.Set("Origin", "http://localhost:5173")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if got, want := response.Code, http.StatusNoContent; got != want {
		t.Fatalf("status code = %d, want %d", got, want)
	}
	if got, want := response.Header().Get("Access-Control-Allow-Origin"), "http://localhost:5173"; got != want {
		t.Fatalf("Access-Control-Allow-Origin = %q, want %q", got, want)
	}
	if got, want := response.Header().Get("Access-Control-Allow-Methods"), "GET, POST, OPTIONS"; got != want {
		t.Fatalf("Access-Control-Allow-Methods = %q, want %q", got, want)
	}
	if got, want := response.Header().Get("Access-Control-Allow-Headers"), "Content-Type"; got != want {
		t.Fatalf("Access-Control-Allow-Headers = %q, want %q", got, want)
	}
}

func TestCORSMiddlewareAllowsLoopbackOrigin(t *testing.T) {
	handler := withCORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called for preflight requests")
	}))

	request := httptest.NewRequest(http.MethodOptions, "/", nil)
	request.Header.Set("Origin", "http://127.0.0.1:3000")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if got, want := response.Code, http.StatusNoContent; got != want {
		t.Fatalf("status code = %d, want %d", got, want)
	}
	if got, want := response.Header().Get("Access-Control-Allow-Origin"), "http://127.0.0.1:3000"; got != want {
		t.Fatalf("Access-Control-Allow-Origin = %q, want %q", got, want)
	}
	if got, want := response.Header().Get("Access-Control-Allow-Methods"), "GET, POST, OPTIONS"; got != want {
		t.Fatalf("Access-Control-Allow-Methods = %q, want %q", got, want)
	}
	if got, want := response.Header().Get("Access-Control-Allow-Headers"), "Content-Type"; got != want {
		t.Fatalf("Access-Control-Allow-Headers = %q, want %q", got, want)
	}
}

func TestCORSMiddlewareRejectsExternalOrigin(t *testing.T) {
	handler := withCORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called for preflight requests")
	}))

	request := httptest.NewRequest(http.MethodOptions, "/", nil)
	request.Header.Set("Origin", "https://example.com")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if got, want := response.Code, http.StatusNoContent; got != want {
		t.Fatalf("status code = %d, want %d", got, want)
	}
	if got := response.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("Access-Control-Allow-Origin = %q, want empty", got)
	}
	if got := response.Header().Get("Access-Control-Allow-Methods"); got != "" {
		t.Fatalf("Access-Control-Allow-Methods = %q, want empty", got)
	}
	if got := response.Header().Get("Access-Control-Allow-Headers"); got != "" {
		t.Fatalf("Access-Control-Allow-Headers = %q, want empty", got)
	}
}

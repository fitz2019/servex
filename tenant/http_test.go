package tenant

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPMiddleware_Success(t *testing.T) {
	resolver := &mockResolver{
		tenant: &testTenant{id: "t1", enabled: true},
	}

	handler := HTTPMiddleware(resolver)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := ID(r.Context())
		w.Write([]byte(id))
	}))

	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Authorization", "Bearer my-token")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if w.Body.String() != "t1" {
		t.Fatalf("body = %q, want %q", w.Body.String(), "t1")
	}
}

func TestHTTPMiddleware_MissingToken(t *testing.T) {
	resolver := &mockResolver{}

	handler := HTTPMiddleware(resolver)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestHTTPMiddleware_ResolveError(t *testing.T) {
	resolver := &mockResolver{err: ErrTenantNotFound}

	handler := HTTPMiddleware(resolver)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Authorization", "Bearer bad-token")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestHTTPMiddleware_Disabled(t *testing.T) {
	resolver := &mockResolver{
		tenant: &testTenant{id: "t1", enabled: false},
	}

	handler := HTTPMiddleware(resolver)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Authorization", "Bearer my-token")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, r)

	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestHTTPMiddleware_Skipper(t *testing.T) {
	resolver := &mockResolver{err: errors.New("should not be called")}

	handler := HTTPMiddleware(resolver,
		WithSkipper(func(_ context.Context, _ any) bool { return true }),
	)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("skipped"))
	}))

	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if w.Body.String() != "skipped" {
		t.Fatalf("body = %q, want %q", w.Body.String(), "skipped")
	}
}

func TestHTTPMiddleware_CustomExtractor(t *testing.T) {
	resolver := &mockResolver{
		tenant: &testTenant{id: "custom-t", enabled: true},
	}

	handler := HTTPMiddleware(resolver,
		WithTokenExtractor(HeaderTokenExtractor("X-Tenant-ID")),
	)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(ID(r.Context())))
	}))

	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("X-Tenant-ID", "custom-t")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if w.Body.String() != "custom-t" {
		t.Fatalf("body = %q, want %q", w.Body.String(), "custom-t")
	}
}

func TestHTTPSkipPaths(t *testing.T) {
	skipper := HTTPSkipPaths("/health", "/api/public/*")

	tests := []struct {
		path string
		want bool
	}{
		{"/health", true},
		{"/api/public/v1", true},
		{"/api/public/v2/foo", true},
		{"/api/private", false},
		{"/healthz", false},
	}

	for _, tt := range tests {
		r, _ := http.NewRequest("GET", tt.path, nil)
		got := skipper(t.Context(), r)
		if got != tt.want {
			t.Errorf("HTTPSkipPaths(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

func TestHTTPSkipPaths_NonHTTP(t *testing.T) {
	skipper := HTTPSkipPaths("/health")
	got := skipper(t.Context(), "not-http")
	if got {
		t.Fatal("非 HTTP 请求应返回 false")
	}
}

func TestHTTPMiddleware_PanicOnNilResolver(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("应在 nil resolver 时 panic")
		}
	}()
	HTTPMiddleware(nil)
}

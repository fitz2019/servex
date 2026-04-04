package cors

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHTTPMiddleware_NonCORSRequest(t *testing.T) {
	handler := HTTPMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
}

func TestHTTPMiddleware_AllowAllOrigins(t *testing.T) {
	handler := HTTPMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestHTTPMiddleware_SpecificOrigin(t *testing.T) {
	handler := HTTPMiddleware(
		WithAllowOrigins("https://example.com", "https://api.example.com"),
	)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	t.Run("允许的来源", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Origin", "https://example.com")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		assert.Equal(t, "https://example.com", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Contains(t, w.Header().Get("Vary"), "Origin")
	})

	t.Run("不允许的来源", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Origin", "https://evil.com")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
	})
}

func TestHTTPMiddleware_Preflight(t *testing.T) {
	handler := HTTPMiddleware(
		WithAllowOrigins("https://example.com"),
		WithAllowMethods("GET", "POST"),
		WithAllowHeaders("Content-Type", "Authorization"),
		WithMaxAge(3600),
	)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("预检请求不应到达 handler")
	}))

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "POST")
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Content-Type")
	assert.Equal(t, "3600", w.Header().Get("Access-Control-Max-Age"))
}

func TestHTTPMiddleware_AllowCredentials(t *testing.T) {
	handler := HTTPMiddleware(
		WithAllowOrigins("https://example.com"),
		WithAllowCredentials(true),
	)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
}

func TestHTTPMiddleware_ExposeHeaders(t *testing.T) {
	handler := HTTPMiddleware(
		WithExposeHeaders("X-Custom-Header", "X-Request-Id"),
	)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	exposed := w.Header().Get("Access-Control-Expose-Headers")
	assert.Contains(t, exposed, "X-Custom-Header")
	assert.Contains(t, exposed, "X-Request-Id")
}

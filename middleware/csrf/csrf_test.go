package csrf

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPMiddleware_SafeMethodSetsCookie(t *testing.T) {
	cfg := DefaultConfig()
	handler := HTTPMiddleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证 token 注入到 context
		token := TokenFromContext(r.Context())
		assert.NotEmpty(t, token)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// 验证 cookie 已设置
	cookies := w.Result().Cookies()
	var csrfCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "_csrf" {
			csrfCookie = c
			break
		}
	}
	require.NotNil(t, csrfCookie, "应设置 CSRF cookie")
	assert.NotEmpty(t, csrfCookie.Value)
	assert.Equal(t, "/", csrfCookie.Path)
}

func TestHTTPMiddleware_UnsafeMethodValidatesToken(t *testing.T) {
	cfg := DefaultConfig()
	handler := HTTPMiddleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	token := "valid-test-token-1234567890abcdef"

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.AddCookie(&http.Cookie{Name: "_csrf", Value: token})
	req.Header.Set("X-CSRF-Token", token)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHTTPMiddleware_MissingTokenReturns403(t *testing.T) {
	cfg := DefaultConfig()
	handler := HTTPMiddleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("不应到达 handler")
	}))

	t.Run("无cookie无header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("有cookie无header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.AddCookie(&http.Cookie{Name: "_csrf", Value: "some-token"})
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestHTTPMiddleware_InvalidTokenReturns403(t *testing.T) {
	cfg := DefaultConfig()
	handler := HTTPMiddleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("不应到达 handler")
	}))

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.AddCookie(&http.Cookie{Name: "_csrf", Value: "cookie-token"})
	req.Header.Set("X-CSRF-Token", "different-token")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestHTTPMiddleware_FormFieldToken(t *testing.T) {
	cfg := DefaultConfig()
	handler := HTTPMiddleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	token := "valid-test-token-1234567890abcdef"

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("csrf_token="+token))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "_csrf", Value: token})
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHTTPMiddleware_SkipperWorks(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Skipper = func(r *http.Request) bool {
		return strings.HasPrefix(r.URL.Path, "/api/")
	}
	handler := HTTPMiddleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// 跳过的路由不需要 token
	req := httptest.NewRequest(http.MethodPost, "/api/data", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHTTPMiddleware_CustomErrorHandler(t *testing.T) {
	cfg := DefaultConfig()
	cfg.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		w.WriteHeader(http.StatusTeapot)
		_, _ = w.Write([]byte("custom error"))
	}
	handler := HTTPMiddleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("不应到达 handler")
	}))

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTeapot, w.Code)
	assert.Contains(t, w.Body.String(), "custom error")
}

func TestHTTPMiddleware_NilConfig(t *testing.T) {
	handler := HTTPMiddleware(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTokenFromContext_Empty(t *testing.T) {
	token := TokenFromContext(httptest.NewRequest(http.MethodGet, "/", nil).Context())
	assert.Empty(t, token)
}

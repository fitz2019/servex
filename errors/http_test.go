package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
)

var (
	testErrTokenExpired = New(100401, "auth.token.expired", "令牌已过期").
				WithHTTP(http.StatusUnauthorized).WithGRPC(codes.Unauthenticated)
	testErrInternal = New(900500, "internal", "服务内部错误").
			WithHTTP(http.StatusInternalServerError).WithGRPC(codes.Internal)
)

func TestToHTTPStatus(t *testing.T) {
	t.Run("from *Error with HTTP set", func(t *testing.T) {
		assert.Equal(t, 401, ToHTTPStatus(testErrTokenExpired))
	})

	t.Run("from *Error without HTTP set", func(t *testing.T) {
		err := New(999, "unknown", "未知错误")
		assert.Equal(t, 500, ToHTTPStatus(err))
	})

	t.Run("from standard error", func(t *testing.T) {
		assert.Equal(t, 500, ToHTTPStatus(fmt.Errorf("plain")))
	})

	t.Run("from nil", func(t *testing.T) {
		assert.Equal(t, 500, ToHTTPStatus(nil))
	})

	t.Run("from wrapped *Error", func(t *testing.T) {
		wrapped := testErrTokenExpired.WithCause(fmt.Errorf("bad"))
		assert.Equal(t, 401, ToHTTPStatus(wrapped))
	})
}

func TestHTTPErrorHandler(t *testing.T) {
	t.Run("handler returns *Error via context", func(t *testing.T) {
		handler := HTTPErrorHandler()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			WriteError(w, testErrTokenExpired)
		}))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))

		var body map[string]any
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		assert.Equal(t, float64(100401), body["code"])
		assert.Equal(t, "auth.token.expired", body["key"])
		assert.Equal(t, "令牌已过期", body["message"])
	})

	t.Run("handler with metadata", func(t *testing.T) {
		errWithMeta := testErrTokenExpired.WithMeta("user_id", "123")
		handler := HTTPErrorHandler()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			WriteError(w, errWithMeta)
		}))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		var body map[string]any
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		meta, ok := body["metadata"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "123", meta["user_id"])
	})

	t.Run("handler success passes through", func(t *testing.T) {
		handler := HTTPErrorHandler()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"ok":true}`))
		}))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, `{"ok":true}`, rec.Body.String())
	})
}

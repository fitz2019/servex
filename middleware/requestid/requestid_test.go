package requestid

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPMiddleware_GeneratesID(t *testing.T) {
	var capturedID string

	handler := HTTPMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, ok := FromContext(r.Context())
		require.True(t, ok)
		capturedID = id
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.NotEmpty(t, capturedID)
	// 响应头应包含相同 ID
	assert.Equal(t, capturedID, w.Header().Get(DefaultHeader))
}

func TestHTTPMiddleware_ReusesExistingID(t *testing.T) {
	const existingID = "my-request-id-123"
	var capturedID string

	handler := HTTPMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, _ := FromContext(r.Context())
		capturedID = id
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(DefaultHeader, existingID)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, existingID, capturedID)
	assert.Equal(t, existingID, w.Header().Get(DefaultHeader))
}

func TestHTTPMiddleware_CustomHeader(t *testing.T) {
	const customHeader = "X-Trace-Id"
	const existingID = "trace-abc"
	var capturedID string

	handler := HTTPMiddleware(WithHeader(customHeader))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, _ := FromContext(r.Context())
		capturedID = id
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(customHeader, existingID)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, existingID, capturedID)
	assert.Equal(t, existingID, w.Header().Get(customHeader))
}

func TestHTTPMiddleware_CustomGenerator(t *testing.T) {
	const fixedID = "fixed-id-001"
	var capturedID string

	handler := HTTPMiddleware(
		WithGenerator(func() string { return fixedID }),
	)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, _ := FromContext(r.Context())
		capturedID = id
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, fixedID, capturedID)
}

func TestFromContext_Empty(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	_, ok := FromContext(req.Context())
	assert.False(t, ok)
}

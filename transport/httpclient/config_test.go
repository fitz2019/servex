package httpclient

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFromConfig_MinimalConfig(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c := NewFromConfig(&Config{
		BaseURL: srv.URL,
		Timeout: 5 * time.Second,
	})
	require.NotNil(t, c)

	resp, err := c.DoRequest(t.Context(), &Request{Method: http.MethodGet, Path: "/api"})
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestNewFromConfig_WithRetry(t *testing.T) {
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(503)
			return
		}
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]any{"ok": true})
	}))
	defer srv.Close()

	c := NewFromConfig(&Config{
		BaseURL:    srv.URL,
		Timeout:    5 * time.Second,
		MaxRetries: 3,
		RetryDelay: time.Millisecond,
	})

	resp, err := c.DoRequest(t.Context(), &Request{Method: http.MethodGet, Path: "/api"})
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, 3, attempts)
}

func TestNewFromConfig_WithAdditionalOpts(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "test-service", r.Header.Get("X-Service"))
		w.WriteHeader(200)
	}))
	defer srv.Close()

	c := NewFromConfig(
		&Config{BaseURL: srv.URL},
		WithHeader("X-Service", "test-service"),
	)

	resp, err := c.DoRequest(t.Context(), &Request{Method: http.MethodGet, Path: "/api"})
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestNewFromConfig_ZeroValues(t *testing.T) {
	c := NewFromConfig(&Config{})
	require.NotNil(t, c)
}

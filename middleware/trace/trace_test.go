package trace

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

func TestHTTPMiddleware_GeneratesTraceID(t *testing.T) {
	var capturedTraceID string

	handler := HTTPMiddleware(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedTraceID = TraceIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// 应自动生成 trace ID
	assert.NotEmpty(t, capturedTraceID)
	// 响应头应包含 trace ID
	assert.Equal(t, capturedTraceID, w.Header().Get("X-Trace-ID"))
}

func TestHTTPMiddleware_PropagatesExisting(t *testing.T) {
	const existingTraceID = "existing-trace-id-abc"
	var capturedTraceID string

	handler := HTTPMiddleware(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedTraceID = TraceIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Trace-ID", existingTraceID)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// 应保留已有的 trace ID
	assert.Equal(t, existingTraceID, capturedTraceID)
	assert.Equal(t, existingTraceID, w.Header().Get("X-Trace-ID"))
}

func TestHTTPMiddleware_SetsResponseHeaders(t *testing.T) {
	handler := HTTPMiddleware(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// 两个响应头都应存在
	assert.NotEmpty(t, w.Header().Get("X-Trace-ID"))
	assert.NotEmpty(t, w.Header().Get("X-Request-ID"))
}

func TestHTTPMiddleware_CustomHeaders(t *testing.T) {
	cfg := &Config{
		TraceIDHeader:   "X-My-Trace",
		RequestIDHeader: "X-My-Request",
	}
	var capturedTraceID string

	handler := HTTPMiddleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedTraceID = TraceIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-My-Trace", "custom-trace")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, "custom-trace", capturedTraceID)
	assert.Equal(t, "custom-trace", w.Header().Get("X-My-Trace"))
	assert.NotEmpty(t, w.Header().Get("X-My-Request"))
}

func TestHTTPMiddleware_InjectsLogger(t *testing.T) {
	// 验证 logger context 中有 traceID（通过 logger.ContextWithTraceID 注入）
	var capturedCtx context.Context

	handler := HTTPMiddleware(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedCtx = r.Context()
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Trace-ID", "logger-trace-123")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// 验证 trace ID 被注入到 context（通过 logger 的 context key）
	traceID := TraceIDFromContext(capturedCtx)
	assert.Equal(t, "logger-trace-123", traceID)
}

func TestInjectHTTPHeaders(t *testing.T) {
	t.Run("注入 trace 和 request ID 到请求头", func(t *testing.T) {
		ctx := withTraceID(context.Background(), "trace-inject-123")
		ctx = withRequestID(ctx, "req-inject-456")

		req := httptest.NewRequest(http.MethodGet, "http://downstream/api", nil)
		InjectHTTPHeaders(ctx, req)

		assert.Equal(t, "trace-inject-123", req.Header.Get("X-Trace-ID"))
		assert.Equal(t, "req-inject-456", req.Header.Get("X-Request-ID"))
	})

	t.Run("无 trace context 时不设置 header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "http://downstream/api", nil)
		InjectHTTPHeaders(context.Background(), req)

		assert.Empty(t, req.Header.Get("X-Trace-ID"))
		assert.Empty(t, req.Header.Get("X-Request-ID"))
	})
}

func TestInjectGRPCMetadata(t *testing.T) {
	t.Run("注入 trace 和 request ID 到 gRPC metadata", func(t *testing.T) {
		ctx := withTraceID(context.Background(), "grpc-trace-123")
		ctx = withRequestID(ctx, "grpc-req-456")

		ctx = InjectGRPCMetadata(ctx)
		md, ok := metadata.FromOutgoingContext(ctx)
		require.True(t, ok)
		assert.Equal(t, []string{"grpc-trace-123"}, md.Get("x-trace-id"))
		assert.Equal(t, []string{"grpc-req-456"}, md.Get("x-request-id"))
	})

	t.Run("无 trace context 时不修改 context", func(t *testing.T) {
		ctx := InjectGRPCMetadata(context.Background())
		_, ok := metadata.FromOutgoingContext(ctx)
		assert.False(t, ok)
	})
}

func TestTraceIDFromContext(t *testing.T) {
	t.Run("存在 trace ID", func(t *testing.T) {
		ctx := withTraceID(context.Background(), "my-trace-id")
		assert.Equal(t, "my-trace-id", TraceIDFromContext(ctx))
	})

	t.Run("不存在 trace ID", func(t *testing.T) {
		assert.Empty(t, TraceIDFromContext(context.Background()))
	})
}

func TestRequestIDFromContext(t *testing.T) {
	t.Run("存在 request ID", func(t *testing.T) {
		ctx := withRequestID(context.Background(), "my-request-id")
		assert.Equal(t, "my-request-id", RequestIDFromContext(ctx))
	})

	t.Run("不存在 request ID", func(t *testing.T) {
		assert.Empty(t, RequestIDFromContext(context.Background()))
	})
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	assert.Equal(t, "X-Trace-ID", cfg.TraceIDHeader)
	assert.Equal(t, "X-Request-ID", cfg.RequestIDHeader)
	assert.Nil(t, cfg.PropagateHeaders)
	assert.Nil(t, cfg.Logger)
}

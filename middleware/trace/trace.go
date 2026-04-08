// Package trace 提供请求链路追踪增强中间件.
//
// 统一 trace-id 在日志、响应头、下游调用中的传播，
// 构建于 middleware/requestid 和 observability/tracing 之上.
package trace

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/Tsukikage7/servex/middleware/requestid"
	"github.com/Tsukikage7/servex/observability/logger"
	"github.com/Tsukikage7/servex/transport/grpcx"
)

// context key 类型，避免与其他包冲突.
type ctxKey int

const (
	traceIDKey ctxKey = iota
	requestIDKey
)

// Config 链路追踪增强配置.
type Config struct {
	// TraceIDHeader HTTP 响应头中的 trace ID 字段名，默认 "X-Trace-ID"
	TraceIDHeader string
	// RequestIDHeader HTTP 响应头中的 request ID 字段名，默认 "X-Request-ID"
	RequestIDHeader string
	// PropagateHeaders 需要传播到下游的 header 列表
	PropagateHeaders []string
	// Logger 日志器（自动注入 trace_id 和 request_id 字段）
	Logger logger.Logger
}

// DefaultConfig 返回默认配置.
func DefaultConfig() *Config {
	return &Config{
		TraceIDHeader:   "X-Trace-ID",
		RequestIDHeader: "X-Request-ID",
	}
}

// applyDefaults 将未设置的字段填充为默认值.
func (c *Config) applyDefaults() {
	if c.TraceIDHeader == "" {
		c.TraceIDHeader = "X-Trace-ID"
	}
	if c.RequestIDHeader == "" {
		c.RequestIDHeader = "X-Request-ID"
	}
}

// generateID 生成新的 UUID.
func generateID() string {
	return uuid.New().String()
}

// HTTPMiddleware HTTP 链路追踪中间件.
// 功能：
//  1. 从请求中提取或生成 trace-id 和 request-id
//  2. 注入到 response header
//  3. 注入到 logger context（后续日志自动带 trace_id）
//  4. 注入到 context 供下游使用
func HTTPMiddleware(cfg *Config) func(http.Handler) http.Handler {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	cfg.applyDefaults()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// 提取或生成 trace-id
			traceID := r.Header.Get(cfg.TraceIDHeader)
			if traceID == "" {
				traceID = generateID()
			}

			// 提取 request-id（优先从 requestid 中间件获取，其次从 header，最后生成）
			reqID, ok := requestid.FromContext(ctx)
			if !ok || reqID == "" {
				reqID = r.Header.Get(cfg.RequestIDHeader)
				if reqID == "" {
					reqID = generateID()
				}
			}

			// 注入到 context
			ctx = withTraceID(ctx, traceID)
			ctx = withRequestID(ctx, reqID)

			// 注入到 logger context（便于后续日志自动带 trace 信息）
			ctx = logger.ContextWithTraceID(ctx, traceID)

			// 设置响应头
			w.Header().Set(cfg.TraceIDHeader, traceID)
			w.Header().Set(cfg.RequestIDHeader, reqID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GRPCUnaryInterceptor gRPC 一元拦截器.
// 从入站 metadata 中提取或生成 trace-id 和 request-id，
// 注入到 context 和响应 metadata.
func GRPCUnaryInterceptor(cfg *Config) grpc.UnaryServerInterceptor {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	cfg.applyDefaults()

	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		ctx = injectTraceContext(ctx, cfg)
		return handler(ctx, req)
	}
}

// GRPCStreamInterceptor gRPC 流式拦截器.
// 从入站 metadata 中提取或生成 trace-id 和 request-id，
// 注入到 context 和响应 metadata.
func GRPCStreamInterceptor(cfg *Config) grpc.StreamServerInterceptor {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	cfg.applyDefaults()

	return func(
		srv any,
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		ctx := injectTraceContext(ss.Context(), cfg)
		wrapped := grpcx.WrapServerStream(ss, ctx)
		return handler(srv, wrapped)
	}
}

// injectTraceContext 从 gRPC metadata 提取或生成 trace/request ID 并注入 context.
func injectTraceContext(ctx context.Context, cfg *Config) context.Context {
	// metadata key 需要小写
	traceKey := grpcMetaKey(cfg.TraceIDHeader)
	reqKey := grpcMetaKey(cfg.RequestIDHeader)

	traceID := grpcx.GetMetadataValue(ctx, traceKey)
	if traceID == "" {
		traceID = generateID()
	}

	reqID, ok := requestid.FromContext(ctx)
	if !ok || reqID == "" {
		reqID = grpcx.GetMetadataValue(ctx, reqKey)
		if reqID == "" {
			reqID = generateID()
		}
	}

	ctx = withTraceID(ctx, traceID)
	ctx = withRequestID(ctx, reqID)
	ctx = logger.ContextWithTraceID(ctx, traceID)

	// 设置响应 metadata
	_ = grpc.SetHeader(ctx, metadata.Pairs(traceKey, traceID, reqKey, reqID))

	return ctx
}

// grpcMetaKey 将 HTTP 风格的 header 名转为 gRPC metadata key（小写）.
func grpcMetaKey(header string) string {
	result := make([]byte, len(header))
	for i := range header {
		c := header[i]
		if c >= 'A' && c <= 'Z' {
			result[i] = c + 32
		} else {
			result[i] = c
		}
	}
	return string(result)
}

// TraceIDFromContext 从 context 获取 trace ID.
func TraceIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(traceIDKey).(string)
	return id
}

// RequestIDFromContext 从 context 获取 request ID.
func RequestIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey).(string)
	return id
}

// InjectHTTPHeaders 将 trace context 注入到 HTTP 请求头（用于客户端调用下游）.
func InjectHTTPHeaders(ctx context.Context, req *http.Request) {
	if traceID := TraceIDFromContext(ctx); traceID != "" {
		req.Header.Set("X-Trace-ID", traceID)
	}
	if reqID := RequestIDFromContext(ctx); reqID != "" {
		req.Header.Set("X-Request-ID", reqID)
	}
}

// InjectGRPCMetadata 将 trace context 注入到 gRPC metadata（用于客户端调用下游）.
func InjectGRPCMetadata(ctx context.Context) context.Context {
	pairs := make([]string, 0, 4)
	if traceID := TraceIDFromContext(ctx); traceID != "" {
		pairs = append(pairs, "x-trace-id", traceID)
	}
	if reqID := RequestIDFromContext(ctx); reqID != "" {
		pairs = append(pairs, "x-request-id", reqID)
	}
	if len(pairs) == 0 {
		return ctx
	}
	return grpcx.AppendOutgoingMetadata(ctx, pairs...)
}

// withTraceID 将 trace ID 存入 context.
func withTraceID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, traceIDKey, id)
}

// withRequestID 将 request ID 存入 context.
func withRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

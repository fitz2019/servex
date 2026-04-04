package requestid

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// UnaryServerInterceptor 创建 gRPC 一元 Request ID 拦截器.
//
// 优先从 gRPC metadata 读取已有 ID，若不存在则生成新 ID，
// 并将 ID 注入 context 和设置响应 metadata 透传.
func UnaryServerInterceptor(opts ...Option) grpc.UnaryServerInterceptor {
	o := defaultOptions(opts)
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		headerKey := canonicalGRPCHeader(o.Header)
		id := extractFromMetadata(ctx, headerKey)
		id = resolveID(id, o.Generator)

		ctx = newContextWithID(ctx, id)

		// 设置响应 metadata 透传
		_ = grpc.SetHeader(ctx, metadata.Pairs(headerKey, id))

		return handler(ctx, req)
	}
}

// extractFromMetadata 从 gRPC incoming metadata 中提取 Request ID.
func extractFromMetadata(ctx context.Context, key string) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	if values := md.Get(key); len(values) > 0 {
		return values[0]
	}
	return ""
}

// canonicalGRPCHeader 将 HTTP 头名转为 gRPC metadata key（小写）.
func canonicalGRPCHeader(header string) string {
	result := make([]byte, len(header))
	for i, c := range header {
		if c >= 'A' && c <= 'Z' {
			result[i] = byte(c + 32)
		} else {
			result[i] = byte(c)
		}
	}
	return string(result)
}

package grpcx

import (
	"context"

	"google.golang.org/grpc"
)

// WrappedServerStream 包装 gRPC ServerStream，允许替换 context.
type WrappedServerStream struct {
	grpc.ServerStream
	Ctx context.Context
}

// Context 返回包装后的 context.
func (w *WrappedServerStream) Context() context.Context {
	return w.Ctx
}

// WrapServerStream 创建包装后的 ServerStream.
func WrapServerStream(stream grpc.ServerStream, ctx context.Context) grpc.ServerStream {
	return &WrappedServerStream{ServerStream: stream, Ctx: ctx}
}

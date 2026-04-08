package grpcx

import (
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Error 创建 gRPC status error.
func Error(code codes.Code, msg string) error {
	return status.Error(code, msg)
}

// Errorf 创建格式化 gRPC status error.
func Errorf(code codes.Code, format string, args ...any) error {
	return status.Error(code, fmt.Sprintf(format, args...))
}

// Code 从 error 中提取 gRPC status code.
// 若 err 为 nil，返回 codes.OK；若无法解析，返回 codes.Unknown.
func Code(err error) codes.Code {
	return status.Code(err)
}

// Message 从 error 中提取 gRPC status message.
// 若 err 为 nil，返回空字符串.
func Message(err error) string {
	if err == nil {
		return ""
	}
	s, _ := status.FromError(err)
	if s == nil {
		return err.Error()
	}
	return s.Message()
}

// IsCode 检查 error 是否匹配指定的 gRPC code.
func IsCode(err error, code codes.Code) bool {
	return status.Code(err) == code
}

// NotFound 创建 codes.NotFound 错误.
func NotFound(msg string) error {
	return status.Error(codes.NotFound, msg)
}

// InvalidArgument 创建 codes.InvalidArgument 错误.
func InvalidArgument(msg string) error {
	return status.Error(codes.InvalidArgument, msg)
}

// PermissionDenied 创建 codes.PermissionDenied 错误.
func PermissionDenied(msg string) error {
	return status.Error(codes.PermissionDenied, msg)
}

// Unauthenticated 创建 codes.Unauthenticated 错误.
func Unauthenticated(msg string) error {
	return status.Error(codes.Unauthenticated, msg)
}

// Internal 创建 codes.Internal 错误.
func Internal(msg string) error {
	return status.Error(codes.Internal, msg)
}

// Unavailable 创建 codes.Unavailable 错误.
func Unavailable(msg string) error {
	return status.Error(codes.Unavailable, msg)
}

// AlreadyExists 创建 codes.AlreadyExists 错误.
func AlreadyExists(msg string) error {
	return status.Error(codes.AlreadyExists, msg)
}

// DeadlineExceeded 创建 codes.DeadlineExceeded 错误.
func DeadlineExceeded(msg string) error {
	return status.Error(codes.DeadlineExceeded, msg)
}

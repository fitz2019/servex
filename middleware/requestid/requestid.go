// Package requestid 提供 Request ID 中间件.
//
// 优先从请求头读取已有 ID，若不存在则用 UUID 生成新 ID，
// 并将 ID 写入 context 和响应头透传.
package requestid

import (
	"context"

	"github.com/google/uuid"
)

// DefaultHeader 默认的 Request ID 请求头名称.
const DefaultHeader = "X-Request-Id"

// contextKey context key 类型.
type contextKey struct{}

// Options Request ID 中间件配置.
type Options struct {
	// Header 读取/写入 Request ID 的请求头名称，默认 DefaultHeader.
	Header string
	// Generator 自定义 ID 生成函数，默认使用 UUID v4.
	Generator func() string
}

// Option 配置函数.
type Option func(*Options)

// WithHeader 设置自定义请求头名称.
func WithHeader(header string) Option {
	return func(o *Options) { o.Header = header }
}

// WithGenerator 设置自定义 ID 生成函数.
func WithGenerator(fn func() string) Option {
	return func(o *Options) { o.Generator = fn }
}

// defaultOptions 默认配置.
func defaultOptions(opts []Option) Options {
	o := Options{
		Header:    DefaultHeader,
		Generator: func() string { return uuid.New().String() },
	}
	for _, opt := range opts {
		opt(&o)
	}
	return o
}

// FromContext 从 context 中读取 Request ID.
func FromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(contextKey{}).(string)
	return id, ok && id != ""
}

// newContextWithID 将 Request ID 存入 context.
func newContextWithID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, contextKey{}, id)
}

// resolveID 从请求头中读取 ID，若不存在则生成新 ID.
func resolveID(existing string, generator func() string) string {
	if existing != "" {
		return existing
	}
	return generator()
}

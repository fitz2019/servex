// Package middleware 提供 AI 模型调用的中间件链.
package middleware

import (
	"context"

	"github.com/Tsukikage7/servex/ai"
)

// Middleware AI 模型中间件，接收一个 ChatModel 返回一个装饰后的 ChatModel.
type Middleware func(ai.ChatModel) ai.ChatModel

// Chain 将多个中间件链接在一起.
// 第一个参数是最外层中间件，最后一个参数是最内层中间件（最先执行）.
func Chain(outer Middleware, others ...Middleware) Middleware {
	return func(next ai.ChatModel) ai.ChatModel {
		for i := len(others) - 1; i >= 0; i-- {
			next = others[i](next)
		}
		return outer(next)
	}
}

// wrappedModel 通过 generate/stream 函数对包装 ChatModel.
type wrappedModel struct {
	generateFn func(ctx context.Context, messages []ai.Message, opts ...ai.CallOption) (*ai.ChatResponse, error)
	streamFn   func(ctx context.Context, messages []ai.Message, opts ...ai.CallOption) (ai.StreamReader, error)
}

// Generate 实现 ai.ChatModel.
func (w *wrappedModel) Generate(ctx context.Context, messages []ai.Message, opts ...ai.CallOption) (*ai.ChatResponse, error) {
	return w.generateFn(ctx, messages, opts...)
}

// Stream 实现 ai.ChatModel.
func (w *wrappedModel) Stream(ctx context.Context, messages []ai.Message, opts ...ai.CallOption) (ai.StreamReader, error) {
	return w.streamFn(ctx, messages, opts...)
}

// 编译期接口断言.
var _ ai.ChatModel = (*wrappedModel)(nil)

// Wrap 将 generate/stream 函数对包装为 ai.ChatModel.
func Wrap(
	generateFn func(ctx context.Context, messages []ai.Message, opts ...ai.CallOption) (*ai.ChatResponse, error),
	streamFn func(ctx context.Context, messages []ai.Message, opts ...ai.CallOption) (ai.StreamReader, error),
) ai.ChatModel {
	return &wrappedModel{
		generateFn: generateFn,
		streamFn:   streamFn,
	}
}

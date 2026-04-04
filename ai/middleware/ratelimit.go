package middleware

import (
	"context"

	"github.com/Tsukikage7/servex/ai"
	"github.com/Tsukikage7/servex/middleware/ratelimit"
)

// RateLimit 返回基于 ratelimit.Limiter 的限流中间件.
// 每次 Generate/Stream 前调用 limiter.Wait，被取消时返回 context 错误.
func RateLimit(limiter ratelimit.Limiter) Middleware {
	return func(next ai.ChatModel) ai.ChatModel {
		return Wrap(
			func(ctx context.Context, messages []ai.Message, opts ...ai.CallOption) (*ai.ChatResponse, error) {
				if err := limiter.Wait(ctx); err != nil {
					return nil, err
				}
				return next.Generate(ctx, messages, opts...)
			},
			func(ctx context.Context, messages []ai.Message, opts ...ai.CallOption) (ai.StreamReader, error) {
				if err := limiter.Wait(ctx); err != nil {
					return nil, err
				}
				return next.Stream(ctx, messages, opts...)
			},
		)
	}
}

package middleware

import (
	"context"
	"time"

	"github.com/Tsukikage7/servex/ai"
)

// Retry 返回对 429/5xx 错误进行指数退避重试的中间件.
//
// maxAttempts: 最大尝试次数（含首次），最少为 1.
// baseDelay: 首次重试等待时间，后续按 2^n 倍增，最大 30 秒.
func Retry(maxAttempts int, baseDelay time.Duration) Middleware {
	if maxAttempts < 1 {
		maxAttempts = 1
	}
	return func(next ai.ChatModel) ai.ChatModel {
		return Wrap(
			func(ctx context.Context, messages []ai.Message, opts ...ai.CallOption) (*ai.ChatResponse, error) {
				var lastErr error
				for attempt := range maxAttempts {
					resp, err := next.Generate(ctx, messages, opts...)
					if err == nil {
						return resp, nil
					}
					lastErr = err
					if !ai.IsRetryable(err) {
						return nil, err
					}
					if attempt == maxAttempts-1 {
						break
					}
					delay := calcDelay(baseDelay, attempt)
					select {
					case <-ctx.Done():
						return nil, ctx.Err()
					case <-time.After(delay):
					}
				}
				return nil, lastErr
			},
			func(ctx context.Context, messages []ai.Message, opts ...ai.CallOption) (ai.StreamReader, error) {
				var lastErr error
				for attempt := range maxAttempts {
					reader, err := next.Stream(ctx, messages, opts...)
					if err == nil {
						return reader, nil
					}
					lastErr = err
					if !ai.IsRetryable(err) {
						return nil, err
					}
					if attempt == maxAttempts-1 {
						break
					}
					delay := calcDelay(baseDelay, attempt)
					select {
					case <-ctx.Done():
						return nil, ctx.Err()
					case <-time.After(delay):
					}
				}
				return nil, lastErr
			},
		)
	}
}

// calcDelay 计算指数退避延迟，最大 30 秒.
func calcDelay(base time.Duration, attempt int) time.Duration {
	delay := base
	for range attempt {
		delay *= 2
	}
	const maxDelay = 30 * time.Second
	if delay > maxDelay {
		delay = maxDelay
	}
	return delay
}

package ratelimit

import "errors"

var (
	// ErrRateLimited 请求被限流.
	ErrRateLimited = errors.New("ratelimit: 请求被限流")

	// ErrNilLimiter 限流器为空.
	ErrNilLimiter = errors.New("ratelimit: 限流器不能为空")

	// ErrInvalidConfig 配置无效.
	ErrInvalidConfig = errors.New("ratelimit: 配置无效")

	// ErrNilCache 缓存为空.
	ErrNilCache = errors.New("ratelimit: 分布式限流需要缓存")
)

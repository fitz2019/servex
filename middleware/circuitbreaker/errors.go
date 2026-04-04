package circuitbreaker

import "errors"

// 预定义错误.
var (
	// ErrCircuitOpen 熔断器开路，请求被拒绝.
	ErrCircuitOpen = errors.New("circuitbreaker: 熔断器已开路，请求被拒绝")
)

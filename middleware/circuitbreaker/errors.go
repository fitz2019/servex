package circuitbreaker

import "errors"

var (
	// ErrCircuitOpen 熔断器开路，请求被拒绝.
	ErrCircuitOpen = errors.New("circuitbreaker: 熔断器已开路，请求被拒绝")
)

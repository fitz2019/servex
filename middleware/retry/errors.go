package retry

import "errors"

var (
	// ErrMaxAttempts 已达到最大重试次数.
	ErrMaxAttempts = errors.New("已达到最大重试次数")
)

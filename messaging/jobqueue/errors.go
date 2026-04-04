// jobqueue/errors.go
package jobqueue

import "errors"

var (
	ErrClosed         = errors.New("jobqueue: 已关闭")
	ErrNilJob         = errors.New("jobqueue: job 为空")
	ErrEmptyQueue     = errors.New("jobqueue: queue 名称为空")
	ErrEmptyType      = errors.New("jobqueue: job type 为空")
	ErrNoHandler      = errors.New("jobqueue: 未找到对应的 handler")
	ErrNoQueues       = errors.New("jobqueue: 未配置要消费的 queue")
	ErrJobNotFound    = errors.New("jobqueue: job 未找到")
	ErrDequeueTimeout = errors.New("jobqueue: 拉取超时")
)

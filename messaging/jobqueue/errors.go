package jobqueue

import "errors"

var (
	// ErrClosed 表示队列已关闭.
	ErrClosed = errors.New("jobqueue: 已关闭")
	// ErrNilJob 表示 job 参数为空.
	ErrNilJob = errors.New("jobqueue: job 为空")
	// ErrEmptyQueue 表示 queue 名称为空.
	ErrEmptyQueue = errors.New("jobqueue: queue 名称为空")
	// ErrEmptyType 表示 job type 为空.
	ErrEmptyType = errors.New("jobqueue: job type 为空")
	// ErrNoHandler 表示未找到对应的 handler.
	ErrNoHandler = errors.New("jobqueue: 未找到对应的 handler")
	// ErrNoQueues 表示未配置要消费的 queue.
	ErrNoQueues = errors.New("jobqueue: 未配置要消费的 queue")
	// ErrJobNotFound 表示 job 未找到.
	ErrJobNotFound = errors.New("jobqueue: job 未找到")
	// ErrDequeueTimeout 表示拉取任务超时.
	ErrDequeueTimeout = errors.New("jobqueue: 拉取超时")
)

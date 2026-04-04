// jobqueue/options.go
package jobqueue

import (
	"time"

	"github.com/Tsukikage7/servex/observability/logger"
)

type workerOptions struct {
	queues       []string
	concurrency  int
	pollInterval time.Duration
	logger       logger.Logger
}

// WorkerOption 配置 Worker。
type WorkerOption func(*workerOptions)

func WithQueues(queues ...string) WorkerOption {
	return func(o *workerOptions) { o.queues = queues }
}

func WithConcurrency(n int) WorkerOption {
	return func(o *workerOptions) { o.concurrency = n }
}

func WithPollInterval(d time.Duration) WorkerOption {
	return func(o *workerOptions) { o.pollInterval = d }
}

func WithLogger(log logger.Logger) WorkerOption {
	return func(o *workerOptions) { o.logger = log }
}

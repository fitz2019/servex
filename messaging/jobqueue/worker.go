package jobqueue

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type worker struct {
	store    Store
	handlers map[string]Handler
	mu       sync.RWMutex
	closed   atomic.Bool
	opts     workerOptions
}

// NewWorker 创建任务消费 Worker.
func NewWorker(store Store, opts ...WorkerOption) Worker {
	o := workerOptions{
		concurrency:  1,
		pollInterval: time.Second,
	}
	for _, opt := range opts {
		opt(&o)
	}
	return &worker{
		store:    store,
		handlers: make(map[string]Handler),
		opts:     o,
	}
}

func (w *worker) Register(jobType string, handler Handler) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.handlers[jobType] = handler
}

// Start 启动 Worker，阻塞直到 ctx 取消.
func (w *worker) Start(ctx context.Context) error {
	if len(w.opts.queues) == 0 {
		return ErrNoQueues
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, w.opts.concurrency)

	for {
		select {
		case <-ctx.Done():
			wg.Wait()
			return nil
		default:
		}

		job := w.fetchJob(ctx)
		if job == nil {
			select {
			case <-ctx.Done():
				wg.Wait()
				return nil
			case <-time.After(w.opts.pollInterval):
				continue
			}
		}

		sem <- struct{}{}
		wg.Go(func() {
			defer func() { <-sem }()
			w.processJob(ctx, job)
		})
	}
}

func (w *worker) fetchJob(ctx context.Context) *Job {
	for _, queue := range w.opts.queues {
		job, err := w.store.Dequeue(ctx, queue)
		if err != nil {
			continue
		}
		return job
	}
	return nil
}

func (w *worker) processJob(ctx context.Context, job *Job) {
	w.store.MarkRunning(ctx, job.ID)

	w.mu.RLock()
	handler, ok := w.handlers[job.Type]
	w.mu.RUnlock()

	if !ok {
		w.store.MarkFailed(ctx, job.ID, ErrNoHandler)
		return
	}

	jobCtx := ctx
	if !job.Deadline.IsZero() {
		var cancel context.CancelFunc
		jobCtx, cancel = context.WithDeadline(ctx, job.Deadline)
		defer cancel()
	}

	if err := handler(jobCtx, job); err != nil {
		job.Retried++
		job.LastError = err.Error()
		if job.Retried >= job.MaxRetries {
			w.store.MarkDead(ctx, job.ID)
		} else {
			job.Status = StatusPending
			job.ScheduledAt = time.Now().Add(w.backoff(job.Retried))
			w.store.Requeue(ctx, job)
		}
		return
	}

	w.store.MarkDone(ctx, job.ID)
}

func (w *worker) backoff(retried int) time.Duration {
	d := 50 * time.Millisecond
	for range retried {
		d *= 2
	}
	if d > 5*time.Minute {
		d = 5 * time.Minute
	}
	return d
}

func (w *worker) Close() error {
	w.closed.Store(true)
	return nil
}

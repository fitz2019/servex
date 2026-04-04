// jobqueue/client.go
package jobqueue

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type client struct {
	store Store
}

// NewClient 创建任务投递客户端。
func NewClient(store Store) Client {
	return &client{store: store}
}

func (c *client) Enqueue(ctx context.Context, job *Job) error {
	if job == nil {
		return ErrNilJob
	}
	if job.Queue == "" {
		return ErrEmptyQueue
	}
	if job.Type == "" {
		return ErrEmptyType
	}

	now := time.Now()
	if job.ID == "" {
		job.ID = uuid.NewString()
	}
	job.Status = StatusPending
	job.CreatedAt = now
	if job.Delay > 0 {
		job.ScheduledAt = now.Add(job.Delay)
	} else {
		job.ScheduledAt = now
	}

	return c.store.Enqueue(ctx, job)
}

func (c *client) Close() error {
	return nil
}

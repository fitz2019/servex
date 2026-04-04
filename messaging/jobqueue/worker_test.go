// jobqueue/worker_test.go
package jobqueue

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type jobStoreForWorker struct {
	mu       sync.Mutex
	jobs     []*Job
	done     []string
	dead     []string
	failed   []string
	requeued []*Job
}

func (s *jobStoreForWorker) Enqueue(_ context.Context, job *Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs = append(s.jobs, job)
	return nil
}

func (s *jobStoreForWorker) Dequeue(_ context.Context, queue string) (*Job, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, j := range s.jobs {
		if j.Queue == queue && j.Status == StatusPending && !j.ScheduledAt.After(time.Now()) {
			s.jobs = append(s.jobs[:i], s.jobs[i+1:]...)
			return j, nil
		}
	}
	return nil, ErrDequeueTimeout
}

func (s *jobStoreForWorker) MarkRunning(_ context.Context, id string) error { return nil }
func (s *jobStoreForWorker) MarkFailed(_ context.Context, id string, _ error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.failed = append(s.failed, id)
	return nil
}
func (s *jobStoreForWorker) MarkDead(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.dead = append(s.dead, id)
	return nil
}
func (s *jobStoreForWorker) MarkDone(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.done = append(s.done, id)
	return nil
}
func (s *jobStoreForWorker) Requeue(_ context.Context, job *Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.requeued = append(s.requeued, job)
	s.jobs = append(s.jobs, job)
	return nil
}
func (s *jobStoreForWorker) ListDead(_ context.Context, _ string) ([]*Job, error) { return nil, nil }
func (s *jobStoreForWorker) Close() error                                         { return nil }

func TestWorker_Register_And_Execute(t *testing.T) {
	store := &jobStoreForWorker{}
	store.Enqueue(t.Context(), &Job{
		ID: "1", Queue: "q", Type: "greet", Status: StatusPending,
		Payload: []byte("hello"), ScheduledAt: time.Now(),
	})

	var called atomic.Bool
	w := NewWorker(store, WithQueues("q"), WithPollInterval(10*time.Millisecond), WithConcurrency(1))
	w.Register("greet", func(ctx context.Context, job *Job) error {
		called.Store(true)
		return nil
	})

	ctx, cancel := context.WithTimeout(t.Context(), 200*time.Millisecond)
	defer cancel()
	w.Start(ctx)

	if !called.Load() {
		t.Error("handler was not called")
	}
	if len(store.done) != 1 || store.done[0] != "1" {
		t.Errorf("expected job 1 to be done, got %v", store.done)
	}
}

func TestWorker_Retry_Then_Dead(t *testing.T) {
	store := &jobStoreForWorker{}
	store.Enqueue(t.Context(), &Job{
		ID: "2", Queue: "q", Type: "fail", Status: StatusPending,
		MaxRetries: 2, ScheduledAt: time.Now(),
	})

	w := NewWorker(store, WithQueues("q"), WithPollInterval(10*time.Millisecond), WithConcurrency(1))
	w.Register("fail", func(ctx context.Context, job *Job) error {
		return errors.New("boom")
	})

	ctx, cancel := context.WithTimeout(t.Context(), 500*time.Millisecond)
	defer cancel()
	w.Start(ctx)

	if len(store.dead) == 0 {
		t.Error("expected job to be dead after retries exhausted")
	}
}

func TestWorker_NoHandler(t *testing.T) {
	store := &jobStoreForWorker{}
	store.Enqueue(t.Context(), &Job{
		ID: "3", Queue: "q", Type: "unknown", Status: StatusPending,
		ScheduledAt: time.Now(),
	})

	w := NewWorker(store, WithQueues("q"), WithPollInterval(10*time.Millisecond), WithConcurrency(1))

	ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
	defer cancel()
	w.Start(ctx)

	if len(store.failed) == 0 {
		t.Error("expected job with no handler to be marked failed")
	}
}

func TestWorker_NoQueues(t *testing.T) {
	w := NewWorker(&jobStoreForWorker{}, WithConcurrency(1))
	err := w.Start(t.Context())
	if !errors.Is(err, ErrNoQueues) {
		t.Errorf("got %v, want ErrNoQueues", err)
	}
}

// jobqueue/client_test.go
package jobqueue

import (
	"context"
	"errors"
	"testing"
	"time"
)

type mockStore struct {
	enqueued []*Job
	err      error
}

func (m *mockStore) Enqueue(_ context.Context, job *Job) error {
	if m.err != nil {
		return m.err
	}
	m.enqueued = append(m.enqueued, job)
	return nil
}

func (m *mockStore) Dequeue(_ context.Context, _ string) (*Job, error)     { return nil, nil }
func (m *mockStore) MarkRunning(_ context.Context, _ string) error         { return nil }
func (m *mockStore) MarkFailed(_ context.Context, _ string, _ error) error { return nil }
func (m *mockStore) MarkDead(_ context.Context, _ string) error            { return nil }
func (m *mockStore) MarkDone(_ context.Context, _ string) error            { return nil }
func (m *mockStore) Requeue(_ context.Context, _ *Job) error               { return nil }
func (m *mockStore) ListDead(_ context.Context, _ string) ([]*Job, error)  { return nil, nil }
func (m *mockStore) Close() error                                          { return nil }

func TestClient_Enqueue(t *testing.T) {
	store := &mockStore{}
	client := NewClient(store)

	job := &Job{
		Queue:   "emails",
		Type:    "welcome",
		Payload: []byte(`{"user":"test"}`),
	}

	err := client.Enqueue(t.Context(), job)
	if err != nil {
		t.Fatal(err)
	}

	if len(store.enqueued) != 1 {
		t.Fatalf("got %d enqueued, want 1", len(store.enqueued))
	}
	if store.enqueued[0].Status != StatusPending {
		t.Errorf("got status %s, want pending", store.enqueued[0].Status)
	}
	if store.enqueued[0].ID == "" {
		t.Error("expected ID to be set")
	}
	if store.enqueued[0].CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

func TestClient_Enqueue_NilJob(t *testing.T) {
	client := NewClient(&mockStore{})
	err := client.Enqueue(t.Context(), nil)
	if !errors.Is(err, ErrNilJob) {
		t.Errorf("got %v, want ErrNilJob", err)
	}
}

func TestClient_Enqueue_EmptyQueue(t *testing.T) {
	client := NewClient(&mockStore{})
	err := client.Enqueue(t.Context(), &Job{Type: "x"})
	if !errors.Is(err, ErrEmptyQueue) {
		t.Errorf("got %v, want ErrEmptyQueue", err)
	}
}

func TestClient_Enqueue_EmptyType(t *testing.T) {
	client := NewClient(&mockStore{})
	err := client.Enqueue(t.Context(), &Job{Queue: "q"})
	if !errors.Is(err, ErrEmptyType) {
		t.Errorf("got %v, want ErrEmptyType", err)
	}
}

func TestClient_Enqueue_WithDelay(t *testing.T) {
	store := &mockStore{}
	client := NewClient(store)

	job := &Job{Queue: "q", Type: "t", Delay: 5 * time.Minute}
	err := client.Enqueue(t.Context(), job)
	if err != nil {
		t.Fatal(err)
	}

	enqueued := store.enqueued[0]
	expected := enqueued.CreatedAt.Add(5 * time.Minute)
	if !enqueued.ScheduledAt.Equal(expected) {
		t.Errorf("ScheduledAt = %v, want %v", enqueued.ScheduledAt, expected)
	}
}

func TestClient_Close(t *testing.T) {
	client := NewClient(&mockStore{})
	if err := client.Close(); err != nil {
		t.Fatal(err)
	}
}

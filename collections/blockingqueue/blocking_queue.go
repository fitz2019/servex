// Package blockingqueue 提供基于环形缓冲区的阻塞队列.
package blockingqueue

import (
	"context"
	"sync"

	"golang.org/x/sync/semaphore"
)

// BlockingQueue 阻塞队列接口.
type BlockingQueue[T any] interface {
	Enqueue(ctx context.Context, item T) error
	Dequeue(ctx context.Context) (T, error)
	Len() int
	IsFull() bool
	IsEmpty() bool
}

// ArrayBlockingQueue 基于环形缓冲区的有界阻塞队列，使用双信号量实现阻塞语义.
type ArrayBlockingQueue[T any] struct {
	data       []T
	head       int
	tail       int
	count      int
	capacity   int
	mu         sync.Mutex
	enqueueCap *semaphore.Weighted
	dequeueCap *semaphore.Weighted
}

// New 创建阻塞队列，capacity 必须大于 0 否则 panic.
func New[T any](capacity int) *ArrayBlockingQueue[T] {
	if capacity <= 0 {
		panic("blockingqueue: capacity must be positive")
	}
	q := &ArrayBlockingQueue[T]{
		data:       make([]T, capacity),
		capacity:   capacity,
		enqueueCap: semaphore.NewWeighted(int64(capacity)),
		dequeueCap: semaphore.NewWeighted(int64(capacity)),
	}
	// dequeueCap 初始化为 0：先占满所有 capacity，表示当前无可出队元素
	q.dequeueCap.TryAcquire(int64(capacity))
	return q
}

// Enqueue 队列满时阻塞，直到有空位或 ctx 取消.
func (q *ArrayBlockingQueue[T]) Enqueue(ctx context.Context, item T) error {
	if err := q.enqueueCap.Acquire(ctx, 1); err != nil {
		return err
	}

	q.mu.Lock()
	q.data[q.tail] = item
	q.tail = (q.tail + 1) % q.capacity
	q.count++
	q.mu.Unlock()

	q.dequeueCap.Release(1)
	return nil
}

// Dequeue 队列空时阻塞，直到有元素或 ctx 取消.
func (q *ArrayBlockingQueue[T]) Dequeue(ctx context.Context) (T, error) {
	if err := q.dequeueCap.Acquire(ctx, 1); err != nil {
		var zero T
		return zero, err
	}

	q.mu.Lock()
	item := q.data[q.head]
	var zero T
	q.data[q.head] = zero // 辅助 GC
	q.head = (q.head + 1) % q.capacity
	q.count--
	q.mu.Unlock()

	q.enqueueCap.Release(1)
	return item, nil
}

func (q *ArrayBlockingQueue[T]) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.count
}

func (q *ArrayBlockingQueue[T]) IsFull() bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.count == q.capacity
}

func (q *ArrayBlockingQueue[T]) IsEmpty() bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.count == 0
}

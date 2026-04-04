// Package delayqueue 提供基于优先队列的延迟队列.
package delayqueue

import (
	"context"
	"sync"
	"time"

	"github.com/Tsukikage7/servex/collections/priorityqueue"
)

// Delayable 可延迟的元素接口.
// Delay 返回距到期的剩余时间，<=0 表示已到期.
type Delayable interface {
	Delay() time.Duration
}

// DelayQueue 延迟队列，元素只有在到期后才能被出队.
type DelayQueue[T Delayable] struct {
	pq     *priorityqueue.PriorityQueue[T]
	mu     sync.Mutex
	signal chan struct{}
}

// New 创建延迟队列，capacity 预留兼容性（当前未使用）.
func New[T Delayable](capacity int) *DelayQueue[T] {
	return &DelayQueue[T]{
		pq: priorityqueue.New(func(a, b T) bool {
			return a.Delay() < b.Delay()
		}),
		signal: make(chan struct{}, 1),
	}
}

// Enqueue 如果新元素成为堆顶（最早到期），会唤醒等待中的 Dequeue.
func (dq *DelayQueue[T]) Enqueue(ctx context.Context, item T) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	dq.mu.Lock()
	oldTop, hadTop := dq.pq.Peek()
	dq.pq.Push(item)
	newTop, _ := dq.pq.Peek()
	shouldSignal := !hadTop || newTop.Delay() < oldTop.Delay()
	dq.mu.Unlock()

	if shouldSignal {
		dq.notify()
	}
	return nil
}

// Dequeue 阻塞直到有元素到期或 ctx 取消.
func (dq *DelayQueue[T]) Dequeue(ctx context.Context) (T, error) {
	for {
		select {
		case <-ctx.Done():
			var zero T
			return zero, ctx.Err()
		default:
		}

		dq.mu.Lock()
		top, ok := dq.pq.Peek()
		if !ok {
			dq.mu.Unlock()
			select {
			case <-dq.signal:
				continue
			case <-ctx.Done():
				var zero T
				return zero, ctx.Err()
			}
		}

		delay := top.Delay()
		if delay <= 0 {
			item, _ := dq.pq.Pop()
			dq.mu.Unlock()
			return item, nil
		}
		dq.mu.Unlock()

		timer := time.NewTimer(delay)
		select {
		case <-timer.C:
			continue
		case <-dq.signal:
			timer.Stop()
			continue
		case <-ctx.Done():
			timer.Stop()
			var zero T
			return zero, ctx.Err()
		}
	}
}

func (dq *DelayQueue[T]) Len() int {
	dq.mu.Lock()
	defer dq.mu.Unlock()
	return dq.pq.Len()
}

func (dq *DelayQueue[T]) IsEmpty() bool {
	dq.mu.Lock()
	defer dq.mu.Unlock()
	return dq.pq.IsEmpty()
}

func (dq *DelayQueue[T]) notify() {
	select {
	case dq.signal <- struct{}{}:
	default:
	}
}

package priorityqueue

import (
	"cmp"
	"sync"
)

// ConcurrentPriorityQueue 线程安全的优先队列.
type ConcurrentPriorityQueue[T any] struct {
	pq *PriorityQueue[T]
	mu sync.RWMutex
}

func NewConcurrent[T any](less LessFunc[T]) *ConcurrentPriorityQueue[T] {
	return &ConcurrentPriorityQueue[T]{pq: New(less)}
}

func NewConcurrentMin[T cmp.Ordered]() *ConcurrentPriorityQueue[T] {
	return &ConcurrentPriorityQueue[T]{pq: NewMin[T]()}
}

func NewConcurrentMax[T cmp.Ordered]() *ConcurrentPriorityQueue[T] {
	return &ConcurrentPriorityQueue[T]{pq: NewMax[T]()}
}

func (cpq *ConcurrentPriorityQueue[T]) Push(items ...T) {
	cpq.mu.Lock()
	defer cpq.mu.Unlock()
	cpq.pq.Push(items...)
}

func (cpq *ConcurrentPriorityQueue[T]) Pop() (T, bool) {
	cpq.mu.Lock()
	defer cpq.mu.Unlock()
	return cpq.pq.Pop()
}

func (cpq *ConcurrentPriorityQueue[T]) Peek() (T, bool) {
	cpq.mu.RLock()
	defer cpq.mu.RUnlock()
	return cpq.pq.Peek()
}

func (cpq *ConcurrentPriorityQueue[T]) Len() int {
	cpq.mu.RLock()
	defer cpq.mu.RUnlock()
	return cpq.pq.Len()
}

func (cpq *ConcurrentPriorityQueue[T]) IsEmpty() bool {
	cpq.mu.RLock()
	defer cpq.mu.RUnlock()
	return cpq.pq.IsEmpty()
}

func (cpq *ConcurrentPriorityQueue[T]) Clear() {
	cpq.mu.Lock()
	defer cpq.mu.Unlock()
	cpq.pq.Clear()
}

// ToSlice 按优先级顺序弹出所有元素，会清空队列.
func (cpq *ConcurrentPriorityQueue[T]) ToSlice() []T {
	cpq.mu.Lock()
	defer cpq.mu.Unlock()
	return cpq.pq.ToSlice()
}

func (cpq *ConcurrentPriorityQueue[T]) Clone() *ConcurrentPriorityQueue[T] {
	cpq.mu.RLock()
	defer cpq.mu.RUnlock()
	return &ConcurrentPriorityQueue[T]{pq: cpq.pq.Clone()}
}

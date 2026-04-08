// Package syncx 提供泛型并发原语.
package syncx

import "sync"

// Pool 泛型对象池，包装 sync.Pool 提供类型安全.
type Pool[T any] struct {
	pool sync.Pool
}

// NewPool 创建泛型对象池，factory 为对象构造函数.
func NewPool[T any](factory func() T) *Pool[T] {
	return &Pool[T]{
		pool: sync.Pool{
			New: func() any {
				return factory()
			},
		},
	}
}

// Get 从池中获取一个对象.
func (p *Pool[T]) Get() T {
	return p.pool.Get().(T)
}

// Put 将对象放回池中.
func (p *Pool[T]) Put(t T) {
	p.pool.Put(t)
}

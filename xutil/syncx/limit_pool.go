package syncx

import (
	"sync"
	"sync/atomic"
)

// LimitPool 带令牌限制的对象池，超过上限时 Get 返回零值和 false.
type LimitPool[T any] struct {
	pool      sync.Pool
	factory   func() T
	tokens    atomic.Int32
	maxTokens int32
}

func NewLimitPool[T any](maxTokens int, factory func() T) *LimitPool[T] {
	lp := &LimitPool[T]{
		factory:   factory,
		maxTokens: int32(maxTokens),
	}
	lp.pool.New = func() any {
		return factory()
	}
	return lp
}

// Get 已达到令牌上限时返回零值和 false.
func (lp *LimitPool[T]) Get() (T, bool) {
	for {
		current := lp.tokens.Load()
		if current >= lp.maxTokens {
			var zero T
			return zero, false
		}
		if lp.tokens.CompareAndSwap(current, current+1) {
			return lp.pool.Get().(T), true
		}
	}
}

// Put 仅当令牌计数大于 0 时才递减，防止未配对调用导致下溢.
func (lp *LimitPool[T]) Put(t T) {
	lp.pool.Put(t)
	for {
		current := lp.tokens.Load()
		if current <= 0 {
			return
		}
		if lp.tokens.CompareAndSwap(current, current-1) {
			return
		}
	}
}

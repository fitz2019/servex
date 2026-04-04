package syncx

import "sync"

// Map 泛型并发安全 Map，包装 sync.Map 提供类型安全.
// 零值可用，无需初始化.
type Map[K comparable, V any] struct {
	m sync.Map
}

func (m *Map[K, V]) Load(key K) (V, bool) {
	val, ok := m.m.Load(key)
	if !ok {
		var zero V
		return zero, false
	}
	return val.(V), true
}

func (m *Map[K, V]) Store(key K, value V) {
	m.m.Store(key, value)
}

func (m *Map[K, V]) LoadOrStore(key K, value V) (V, bool) {
	actual, loaded := m.m.LoadOrStore(key, value)
	return actual.(V), loaded
}

func (m *Map[K, V]) LoadAndDelete(key K) (V, bool) {
	val, loaded := m.m.LoadAndDelete(key)
	if !loaded {
		var zero V
		return zero, false
	}
	return val.(V), true
}

func (m *Map[K, V]) Delete(key K) {
	m.m.Delete(key)
}

func (m *Map[K, V]) Range(fn func(key K, value V) bool) {
	m.m.Range(func(key, value any) bool {
		return fn(key.(K), value.(V))
	})
}

// LoadOrStoreFunc 懒初始化：key 存在时直接返回已有值（loaded=true）；
// 不存在则调用 fn 创建后存入并返回（loaded=false）.
// 返回的 err 来自 fn；若 fn 返回 error，则不存储.
func (m *Map[K, V]) LoadOrStoreFunc(key K, fn func() (V, error)) (actual V, loaded bool, err error) {
	if val, ok := m.Load(key); ok {
		return val, true, nil
	}
	v, err := fn()
	if err != nil {
		return v, false, err
	}
	// 使用 LoadOrStore 处理并发场景：若两个 goroutine 同时到达此处，
	// 只有第一个会成功存储，另一个拿到已存储的值
	actual, loaded = m.LoadOrStore(key, v)
	return actual, loaded, nil
}

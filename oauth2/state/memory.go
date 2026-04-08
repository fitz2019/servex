// Package state 提供 OAuth2 state 参数的生成与验证实现.
package state

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

// MemoryStore 内存实现的 StateStore，用于开发和测试.
type MemoryStore struct {
	mu     sync.Mutex
	states map[string]time.Time
	ttl    time.Duration
}

// NewMemoryStore 创建基于内存的 StateStore 实例.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		states: make(map[string]time.Time),
		ttl:    10 * time.Minute,
	}
}

// Generate 生成一个新的 state 参数并存储.
func (s *MemoryStore) Generate(_ context.Context) (string, error) {
	state := uuid.NewString()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.states[state] = time.Now().Add(s.ttl)
	return state, nil
}

// Validate 验证并消费一个 state 参数.
func (s *MemoryStore) Validate(_ context.Context, state string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	expiresAt, ok := s.states[state]
	if !ok {
		return false, nil
	}
	delete(s.states, state) // 一次性消费
	if time.Now().After(expiresAt) {
		return false, nil
	}
	return true, nil
}

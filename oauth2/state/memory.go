// oauth2/state/memory.go
package state

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

// MemoryStore 内存实现的 StateStore，用于开发和测试。
type MemoryStore struct {
	mu     sync.Mutex
	states map[string]time.Time
	ttl    time.Duration
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		states: make(map[string]time.Time),
		ttl:    10 * time.Minute,
	}
}

func (s *MemoryStore) Generate(_ context.Context) (string, error) {
	state := uuid.NewString()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.states[state] = time.Now().Add(s.ttl)
	return state, nil
}

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

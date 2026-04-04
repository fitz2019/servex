// webhook/store/memory/store.go
package memory

import (
	"context"
	"sync"

	"github.com/Tsukikage7/servex/notify/webhook"
)

// Store 内存实现的 SubscriptionStore，用于开发和测试。
type Store struct {
	mu   sync.RWMutex
	subs map[string]*webhook.Subscription
}

func NewStore() *Store {
	return &Store{subs: make(map[string]*webhook.Subscription)}
}

func (s *Store) Save(_ context.Context, sub *webhook.Subscription) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.subs[sub.ID] = sub
	return nil
}

func (s *Store) Delete(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.subs, id)
	return nil
}

func (s *Store) ListByEvent(_ context.Context, eventType string) ([]*webhook.Subscription, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*webhook.Subscription
	for _, sub := range s.subs {
		if len(sub.Events) == 0 {
			result = append(result, sub)
			continue
		}
		for _, e := range sub.Events {
			if e == eventType {
				result = append(result, sub)
				break
			}
		}
	}
	return result, nil
}

func (s *Store) Get(_ context.Context, id string) (*webhook.Subscription, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sub, ok := s.subs[id]
	if !ok {
		return nil, webhook.ErrNotFound
	}
	return sub, nil
}

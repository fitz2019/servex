// oauth2/state/redis.go
package state

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/Tsukikage7/servex/storage/cache"
)

// RedisStore 基于 Redis 的 StateStore，用于生产环境。
// 通过 servex 的 cache.Cache 接口操作，支持 Redis 和内存缓存。
type RedisStore struct {
	cache  cache.Cache
	prefix string
	ttl    time.Duration
}

type RedisOption func(*RedisStore)

func WithPrefix(prefix string) RedisOption {
	return func(s *RedisStore) { s.prefix = prefix }
}

func WithTTL(ttl time.Duration) RedisOption {
	return func(s *RedisStore) { s.ttl = ttl }
}

// NewRedisStore 创建基于缓存的 StateStore。
// 接受 servex 的 cache.Cache（Redis 或内存均可），复用已有连接。
func NewRedisStore(c cache.Cache, opts ...RedisOption) (*RedisStore, error) {
	if c == nil {
		return nil, errors.New("oauth2/state: cache 不能为空")
	}
	s := &RedisStore{
		cache:  c,
		prefix: "oauth2:state:",
		ttl:    10 * time.Minute,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s, nil
}

func (s *RedisStore) Generate(ctx context.Context) (string, error) {
	state := uuid.NewString()
	if err := s.cache.Set(ctx, s.prefix+state, "1", s.ttl); err != nil {
		return "", err
	}
	return state, nil
}

func (s *RedisStore) Validate(ctx context.Context, state string) (bool, error) {
	key := s.prefix + state
	val, err := s.cache.Get(ctx, key)
	if err != nil {
		if errors.Is(err, cache.ErrNotFound) {
			return false, nil
		}
		return false, err
	}
	// 一次性消费：验证后立即删除
	s.cache.Del(ctx, key)
	return val == "1", nil
}

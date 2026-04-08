package redis

import (
	goredis "github.com/redis/go-redis/v9"
)

// NewStoreFromConfig 根据连接参数创建 Redis Store.
func NewStoreFromConfig(addr, password string, db int, prefix string) (*Store, error) {
	client := goredis.NewClient(&goredis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	var opts []Option
	if prefix != "" {
		opts = append(opts, WithPrefix(prefix))
	}
	return NewStore(client, opts...)
}

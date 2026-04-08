package redis

type options struct {
	prefix string
}

// Option 配置 Redis Store.
type Option func(*options)

// WithPrefix 设置 Redis key 前缀.
func WithPrefix(prefix string) Option {
	return func(o *options) { o.prefix = prefix }
}

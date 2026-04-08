package rabbitmq

type options struct {
	prefix        string
	durable       bool
	prefetchCount int
}

// Option 配置 RabbitMQ Store.
type Option func(*options)

// WithPrefix 设置队列名称前缀.
func WithPrefix(prefix string) Option {
	return func(o *options) { o.prefix = prefix }
}

// WithDurable 设置队列是否持久化.
func WithDurable(d bool) Option {
	return func(o *options) { o.durable = d }
}

// WithPrefetchCount 设置预取消息数量.
func WithPrefetchCount(n int) Option {
	return func(o *options) { o.prefetchCount = n }
}

package kafka

type options struct {
	prefix string
}

// Option 配置 Kafka Store.
type Option func(*options)

// WithPrefix 设置 topic 名称前缀.
func WithPrefix(prefix string) Option {
	return func(o *options) { o.prefix = prefix }
}

// jobqueue/rabbitmq/options.go
package rabbitmq

type options struct {
	prefix        string
	durable       bool
	prefetchCount int
}

type Option func(*options)

func WithPrefix(prefix string) Option {
	return func(o *options) { o.prefix = prefix }
}

func WithDurable(d bool) Option {
	return func(o *options) { o.durable = d }
}

func WithPrefetchCount(n int) Option {
	return func(o *options) { o.prefetchCount = n }
}

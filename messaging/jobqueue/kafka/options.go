// jobqueue/kafka/options.go
package kafka

type options struct {
	prefix string
}

type Option func(*options)

func WithPrefix(prefix string) Option {
	return func(o *options) { o.prefix = prefix }
}

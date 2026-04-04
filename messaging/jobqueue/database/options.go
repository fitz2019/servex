// jobqueue/database/options.go
package database

type options struct {
	tableName string
}

type Option func(*options)

func WithTableName(name string) Option {
	return func(o *options) { o.tableName = name }
}

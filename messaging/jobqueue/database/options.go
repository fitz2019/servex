package database

type options struct {
	tableName string
}

// Option 配置 database Store.
type Option func(*options)

// WithTableName 设置任务表名.
func WithTableName(name string) Option {
	return func(o *options) { o.tableName = name }
}

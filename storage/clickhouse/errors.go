package clickhouse

import "errors"

// 预定义错误.
var (
	// ErrNilConfig 配置为 nil 时返回.
	ErrNilConfig = errors.New("clickhouse: config is nil")
	// ErrNilLogger 日志记录器为 nil 时返回.
	ErrNilLogger = errors.New("clickhouse: logger is nil")
	// ErrEmptyAddrs 地址列表为空时返回.
	ErrEmptyAddrs = errors.New("clickhouse: addrs is empty")
	// ErrEmptyDatabase 数据库名为空时返回.
	ErrEmptyDatabase = errors.New("clickhouse: database name is empty")
)

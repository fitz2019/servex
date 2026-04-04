package clickhouse

import "errors"

// 预定义错误.
var (
	ErrNilConfig     = errors.New("clickhouse: config is nil")
	ErrNilLogger     = errors.New("clickhouse: logger is nil")
	ErrEmptyAddrs    = errors.New("clickhouse: addrs is empty")
	ErrEmptyDatabase = errors.New("clickhouse: database name is empty")
)

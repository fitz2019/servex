package graphql

import "errors"

var (
	// ErrNilSchema 表示传入的 schema 为 nil.
	ErrNilSchema = errors.New("graphql: schema is nil")
	// ErrInvalidRequest 表示请求格式无效.
	ErrInvalidRequest = errors.New("graphql: invalid request")
	// ErrEmptyQuery 表示查询字符串为空.
	ErrEmptyQuery = errors.New("graphql: empty query")
)

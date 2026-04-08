package pbjson

import "errors"

var (
	// ErrNotProtoMessage 不是 proto.Message 类型.
	ErrNotProtoMessage = errors.New("不是 proto.Message 类型")
)

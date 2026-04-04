// Package copier 提供基于反射的结构体对象复制功能.
package copier

import "errors"

var (
	ErrNilSource      = errors.New("copier: 源对象为 nil")
	ErrNilDestination = errors.New("copier: 目标对象为 nil")
	ErrNotStruct      = errors.New("copier: 参数不是结构体类型")
)

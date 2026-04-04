// Package ptrx 提供泛型指针工具函数.
package ptrx

// ToPtr 将值转换为指针.
func ToPtr[T any](v T) *T {
	return &v
}

// ToPtrSlice 将值切片转换为指针切片.
func ToPtrSlice[T any](src []T) []*T {
	dst := make([]*T, len(src))
	for i, v := range src {
		dst[i] = ToPtr(v)
	}
	return dst
}

// Value 获取指针的值，如果指针为 nil 则返回零值.
func Value[T any](ptr *T) T {
	if ptr == nil {
		var zero T
		return zero
	}
	return *ptr
}

// Equal 比较两个指针的值是否相等.
func Equal[T comparable](a, b *T) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

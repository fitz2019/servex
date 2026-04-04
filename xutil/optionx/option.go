// Package optionx 提供泛型函数选项模式的通用实现.
package optionx

type Option[T any] func(*T)

type OptionErr[T any] func(*T) error

func Apply[T any](t *T, opts ...Option[T]) {
	for _, opt := range opts {
		opt(t)
	}
}

// ApplyErr 遇到第一个错误立即返回（fail-fast）.
func ApplyErr[T any](t *T, opts ...OptionErr[T]) error {
	for _, opt := range opts {
		if err := opt(t); err != nil {
			return err
		}
	}
	return nil
}

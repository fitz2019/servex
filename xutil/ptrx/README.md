# ptrx

`ptrx` 提供泛型指针工具函数，简化 Go 中常见的指针操作。

## 功能特性

- `ToPtr` -- 将任意值转换为指针
- `ToPtrSlice` -- 将值切片转换为指针切片
- `Value` -- 安全解引用指针，nil 返回零值
- `Equal` -- 比较两个指针指向的值是否相等

## API

| 函数 | 签名 | 说明 |
| --- | --- | --- |
| `ToPtr` | `ToPtr[T any](v T) *T` | 值转指针 |
| `ToPtrSlice` | `ToPtrSlice[T any](src []T) []*T` | 切片值转指针切片 |
| `Value` | `Value[T any](ptr *T) T` | 安全解引用，nil 返回零值 |
| `Equal` | `Equal[T comparable](a, b *T) bool` | 比较两个指针的值是否相等 |

`Equal` 的判定规则：两者均为 nil 返回 true；仅一方为 nil 返回 false；否则比较解引用后的值。

## 许可证

Apache-2.0

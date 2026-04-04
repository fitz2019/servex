# optionx

`optionx` 提供泛型函数选项模式 (Functional Options Pattern) 的通用实现。

## 功能特性

- `Option[T]` -- 无错误的函数选项类型
- `OptionErr[T]` -- 可返回错误的函数选项类型
- `Apply` -- 将选项依次应用到目标对象
- `ApplyErr` -- 应用可出错的选项，遇到首个错误立即返回 (fail-fast)

## API

### 类型

| 类型 | 定义 | 说明 |
| --- | --- | --- |
| `Option[T]` | `func(*T)` | 无错误的函数选项 |
| `OptionErr[T]` | `func(*T) error` | 可返回错误的函数选项 |

### 函数

| 函数 | 签名 | 说明 |
| --- | --- | --- |
| `Apply` | `Apply[T any](t *T, opts ...Option[T])` | 将选项依次应用到 `t` |
| `ApplyErr` | `ApplyErr[T any](t *T, opts ...OptionErr[T]) error` | 应用选项，遇到错误立即返回 |

## 许可证

Apache-2.0

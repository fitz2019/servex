# valuex

`valuex` 提供类型安全的 `any` 值转换工具，支持精确类型断言、宽松类型转换和默认值回退。

## 功能特性

- `AnyValue` 包装器 -- 封装 `any` 值并提供链式类型访问
- 精确类型断言 -- `Int`、`String`、`Bool` 等，类型不匹配返回错误
- 宽松类型转换 -- `AsInt`、`AsFloat64`、`AsString` 等，自动在数值/字符串间转换
- 默认值方法 -- `IntOrDefault`、`StringOrDefault` 等，转换失败返回默认值
- 语义清晰的错误 -- `ErrNilValue`、`ErrTypeMismatch`、`ErrConvertFailed`

## API

### 构造

| 函数 | 签名 | 说明 |
| --- | --- | --- |
| `Of` | `Of(val any) AnyValue` | 包装 any 值为 AnyValue |

### 精确类型断言

类型必须完全匹配，否则返回 `ErrTypeMismatch`。

`Int`, `Int8`, `Int16`, `Int32`, `Int64`, `Uint`, `Uint8`, `Uint16`, `Uint32`, `Uint64`, `Float32`, `Float64`, `String`, `Bool`, `Bytes`

### 宽松类型转换

自动在数值类型和字符串之间进行转换。

| 方法 | 支持的源类型 |
| --- | --- |
| `AsInt() (int, error)` | 所有整数/浮点类型、string |
| `AsInt64() (int64, error)` | 所有整数/浮点类型、string |
| `AsFloat64() (float64, error)` | 所有整数/浮点类型、string |
| `AsString() (string, error)` | string、[]byte、fmt.Stringer、其他 (fmt.Sprintf) |
| `AsBool() (bool, error)` | bool、int、int64、float64、string |

### 默认值方法

转换失败时返回指定的默认值，不返回错误。

`IntOrDefault`, `Int64OrDefault`, `Float64OrDefault`, `StringOrDefault`, `BoolOrDefault`

### 错误

| 变量 | 说明 |
| --- | --- |
| `ErrNilValue` | 值为 nil |
| `ErrTypeMismatch` | 精确断言类型不匹配 |
| `ErrConvertFailed` | 宽松转换失败 |

## 许可证

Apache-2.0

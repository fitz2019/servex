# strx

`strx` 提供常用的字符串工具函数。

## 功能特性

- 姓名处理：`SplitName`、`JoinName`
- 大小写转换：`TrimAndLower`、`TrimAndUpper`、`ToTitle`
- 空值判断：`IsEmpty`、`IsNotEmpty`、`DefaultIfEmpty`
- 截断：`Truncate`（超长部分用省略号替代）
- 零拷贝转换：`UnsafeToBytes`、`UnsafeToString`

## API

| 函数 | 签名 | 说明 |
| --- | --- | --- |
| `SplitName` | `SplitName(fullName string) (string, string)` | 按首个空格拆分为 firstName 和 lastName |
| `JoinName` | `JoinName(firstName, lastName string) string` | 合并姓名，自动处理空值 |
| `TrimAndLower` | `TrimAndLower(s string) string` | 去除前后空格并转为小写 |
| `TrimAndUpper` | `TrimAndUpper(s string) string` | 去除前后空格并转为大写 |
| `IsEmpty` | `IsEmpty(s string) bool` | 仅含空白字符视为空 |
| `IsNotEmpty` | `IsNotEmpty(s string) bool` | IsEmpty 的取反 |
| `ToTitle` | `ToTitle(s string) string` | 首字母大写，其余小写 |
| `Truncate` | `Truncate(s string, maxLen int) string` | 截断并添加省略号 |
| `DefaultIfEmpty` | `DefaultIfEmpty(s, defaultValue string) string` | 空值时返回默认值 |
| `UnsafeToBytes` | `UnsafeToBytes(s string) []byte` | 零分配转 []byte，返回值不可修改 |
| `UnsafeToString` | `UnsafeToString(b []byte) string` | 零分配转 string，原 []byte 不可再修改 |

> **注意**：`UnsafeToBytes` 和 `UnsafeToString` 使用 `unsafe` 包实现零拷贝转换，
> 修改返回值或原始数据将导致未定义行为。仅在性能敏感场景下使用。

## 许可证

Apache-2.0

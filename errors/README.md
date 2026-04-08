# errors

## 导入路径

```go
import "github.com/Tsukikage7/servex/errors"
```

## 简介

`errors` 包提供统一的业务错误类型 `Error`，包含业务码（Code）、键（Key）、消息（Message）以及可选的 HTTP 状态码和 gRPC Code 映射。支持错误链（`WithCause`）、元数据附加（`WithMeta`）和 `errors.Is` 按 Code 比较。内置 HTTP 响应写入和 gRPC Status 转换工具。

## 核心类型

| 类型 / 函数 | 说明 |
|---|---|
| `Error` | 统一业务错误类型 |
| `New(code, key, message)` | 创建错误定义（通常作为包级变量） |
| `WithHTTP(status)` | 绑定 HTTP 状态码 |
| `WithGRPC(code)` | 绑定 gRPC Code |
| `WithCause(err)` | 包装底层错误（返回新实例） |
| `WithMeta(key, value)` | 附加元数据（返回新实例） |
| `WithMessage(msg)` | 覆盖消息（返回新实例） |
| `FromError(err)` | 从 error 提取 `*Error` |
| `CodeIs(err, target)` | 按 Code 判断错误 |
| `WriteError(w, err)` | 将 `*Error` 写入 HTTP 响应（JSON） |
| `WriteErrorFrom(w, err)` | 将 error 写入 HTTP 响应 |
| `ToGRPCStatus(err)` | 转为 gRPC Status |

## 示例

```go
package main

import (
    "fmt"
    "net/http"

    "google.golang.org/grpc/codes"

    "github.com/Tsukikage7/servex/errors"
)

// 定义错误（通常在包级别）
var (
    ErrUserNotFound = errors.New(404001, "user_not_found", "用户不存在").
        WithHTTP(http.StatusNotFound).
        WithGRPC(codes.NotFound)

    ErrUnauthorized = errors.New(401001, "unauthorized", "未授权").
        WithHTTP(http.StatusUnauthorized)
)

func main() {
    // 附加原始错误
    err := ErrUserNotFound.WithCause(fmt.Errorf("db: record not found"))
    fmt.Println(err) // [404001] user_not_found: 用户不存在: db: record not found

    // 附加元数据
    err2 := ErrUserNotFound.WithMeta("user_id", "u-123")

    // 按 Code 比较（errors.Is 兼容）
    fmt.Println(errors.CodeIs(err2, ErrUserNotFound)) // true

    // 从 error 中提取
    e, ok := errors.FromError(err2)
    if ok {
        fmt.Println("业务码:", e.Code)    // 404001
        fmt.Println("HTTP码:", e.HTTP)    // 404
        fmt.Println("元数据:", e.Metadata) // map[user_id:u-123]
    }

    // 写入 HTTP 响应
    mux := http.NewServeMux()
    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        errors.WriteError(w, ErrUserNotFound)
    })
}
```

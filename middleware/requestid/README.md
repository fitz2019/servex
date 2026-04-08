# middleware/requestid

## 导入路径

```go
import "github.com/Tsukikage7/servex/middleware/requestid"
```

## 简介

`middleware/requestid` 提供 HTTP 和 gRPC 请求 ID 注入中间件。从请求头（HTTP）或 metadata（gRPC）中读取请求 ID，若不存在则自动生成，并写入响应头和上下文，便于链路追踪。默认使用 `X-Request-Id` 头。

## 核心函数

| 函数 | 说明 |
|---|---|
| `HTTPMiddleware(opts...)` | HTTP 请求 ID 中间件 |
| `UnaryServerInterceptor(opts...)` | gRPC 一元服务器拦截器 |
| `WithHeader(header)` | 设置自定义请求 ID 头名（默认 `X-Request-Id`） |
| `WithGenerator(fn)` | 设置自定义 ID 生成函数（默认 UUID） |
| `DefaultHeader` | 默认头名常量 `"X-Request-Id"` |

## 示例

```go
package main

import (
    "context"
    "fmt"
    "net/http"

    "google.golang.org/grpc"

    "github.com/Tsukikage7/servex/middleware/requestid"
)

func main() {
    // HTTP 请求 ID 中间件（默认配置）
    mux := http.NewServeMux()
    mux.HandleFunc("/api/orders", func(w http.ResponseWriter, r *http.Request) {
        // 从上下文中获取请求 ID
        rid := r.Header.Get(requestid.DefaultHeader)
        fmt.Fprintf(w, `{"request_id":"%s"}`, rid)
    })

    handler := requestid.HTTPMiddleware()(mux)

    // 使用自定义头名和生成器
    customHandler := requestid.HTTPMiddleware(
        requestid.WithHeader("X-Trace-ID"),
        requestid.WithGenerator(func() string {
            return fmt.Sprintf("trace-%d", 12345) // 实际使用 UUID 或雪花算法
        }),
    )(mux)
    _ = customHandler

    go http.ListenAndServe(":8080", handler)

    // gRPC 服务器拦截器
    srv := grpc.NewServer(
        grpc.ChainUnaryInterceptor(
            requestid.UnaryServerInterceptor(
                requestid.WithHeader("x-request-id"),
            ),
        ),
    )
    _ = srv

    // 从 context 中读取请求 ID（内部使用）
    ctx := context.Background()
    _ = ctx
    // 请求 ID 会注入到 ctx 中，可通过 ctx.Value 等方式读取
}
```

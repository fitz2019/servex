---
name: middleware
description: servex 中间件模块专家。当用户使用 servex 的 ratelimit、circuitbreaker、retry、recovery、timeout、cors、requestid、idempotency、semaphore、logging 中间件时触发。
---

# servex 中间件

**组合顺序：** requestid → logging → tracing → metrics → ratelimit → circuitbreaker → retry → timeout → recovery

（logging 在 tracing 之前是 servex 约定：tracing 中间件将 trace ID 写入 context，logging 在其后可提取并输出到日志）

## circuitbreaker + ratelimit 组合示例

```go
// 熔断器：连续 5 次失败打开，30s 后进入半开尝试恢复
cb := circuitbreaker.New(
    circuitbreaker.WithFailureThreshold(5),
    circuitbreaker.WithSuccessThreshold(2),
    circuitbreaker.WithOpenTimeout(30 * time.Second),
)

// 令牌桶：100 req/s，桶容量 200（允许瞬时突发）
limiter := ratelimit.NewTokenBucket(100, 200)

// 注意中间件顺序：ratelimit 在 circuitbreaker 之前
srv := httpserver.New(mux,
    httpserver.WithLogger(log),
    httpserver.WithMiddlewares(
        ratelimit.HTTPMiddleware(limiter),
        circuitbreaker.HTTPMiddleware(cb),
    ),
)
```

完整示例：`docs/superpowers/examples/middleware/main.go`

## circuitbreaker — 熔断器

**关键选项：**
- `WithFailureThreshold(n)` — 连续失败 n 次后打开
- `WithSuccessThreshold(n)` — 半开状态成功 n 次后关闭
- `WithOpenTimeout(d)` — Open 状态持续时间，之后进入 HalfOpen

**集成方式：**
- `circuitbreaker.HTTPMiddleware(cb)` — HTTP 中间件（返回 503）
- `circuitbreaker.EndpointMiddleware(cb)` — endpoint 层中间件
- `cb.Execute(ctx, fn)` — 手动执行，自定义错误处理

## ratelimit — 限流

```go
// 令牌桶：平滑限流，允许瞬时突发
limiter := ratelimit.NewTokenBucket(rate, capacity)

// 滑动窗口：精确计数
limiter := ratelimit.NewSlidingWindow(limit, window)

// 固定窗口：性能最好
limiter := ratelimit.NewFixedWindow(limit, window)

// HTTP 中间件（超限返回 429）
ratelimit.HTTPMiddleware(limiter)

// Endpoint 中间件
ratelimit.EndpointMiddleware(limiter)
```

## retry — 重试

```go
// 指数退避重试，最多 3 次，基础间隔 100ms
mw := retry.New(
    retry.WithMaxAttempts(3),
    retry.WithBackoff(retry.ExponentialBackoff(100*time.Millisecond)),
    retry.WithRetryOn(func(err error) bool {
        return errors.Is(err, io.ErrTemporary)
    }),
)
```

## timeout — 超时控制

```go
mw := timeout.New(timeout.WithTimeout(5 * time.Second))
// 超时后返回 504，并取消下游 context
```

## cors — 跨域

```go
mw := cors.New(
    cors.WithAllowOrigins("https://example.com", "https://app.example.com"),
    cors.WithAllowMethods("GET", "POST", "PUT", "DELETE"),
    cors.WithAllowHeaders("Authorization", "Content-Type"),
    cors.WithMaxAge(86400),
)
```

## requestid — 请求 ID

```go
mw := requestid.New() // 自动生成 UUID，写入 X-Request-ID header 和 context

// 在 handler 中读取
id := requestid.FromContext(ctx)
```

## logging — 结构化日志

```go
// HTTP 访问日志
mw := logging.NewHTTP(log, logging.WithSkipPaths("/healthz", "/metrics"))

// gRPC 访问日志
mw := logging.NewGRPC(log)
```

## idempotency — 幂等性

```go
// 基于请求 ID 去重，需要 Store 实现（Redis 或内存）
mw := idempotency.New(
    idempotency.WithStore(redisStore),
    idempotency.WithTTL(24 * time.Hour),
)
```

## semaphore — 并发控制

```go
// 最多 100 个并发请求，超出返回 503
mw := semaphore.New(semaphore.WithLimit(100))
```

## recovery — Panic 恢复

```go
// HTTP 中间件（panic 时返回 500）
httpMw := recovery.HTTPMiddleware(recovery.WithLogger(log))

// gRPC 拦截器（panic 时返回 codes.Internal）
unaryInterceptor := recovery.UnaryServerInterceptor(recovery.WithLogger(log))
streamInterceptor := recovery.StreamServerInterceptor(recovery.WithLogger(log))

// Endpoint 中间件
endpointMw := recovery.EndpointMiddleware(recovery.WithLogger(log))

// 自定义 panic 处理
mw := recovery.HTTPMiddleware(
    recovery.WithLogger(log),
    recovery.WithHandler(func(ctx any, p any, stack []byte) error {
        // 自定义处理逻辑
        return fmt.Errorf("panic recovered: %v", p)
    }),
    recovery.WithStackSize(64 * 1024), // 堆栈大小，默认 64KB
)
```

**关键选项：**
- `recovery.WithLogger(log)` — 日志记录器（必需）
- `recovery.WithHandler(fn)` — 自定义 panic 处理函数
- `recovery.WithStackSize(n)` — 堆栈捕获大小
- 支持三种集成：`HTTPMiddleware`、`UnaryServerInterceptor`/`StreamServerInterceptor`、`EndpointMiddleware`

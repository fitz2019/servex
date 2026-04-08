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

## secure — 安全头

```go
import "github.com/Tsukikage7/servex/middleware/secure"

// 使用默认配置（生产推荐）：自动设置 X-Frame-Options、HSTS、X-Content-Type-Options 等
mw := secure.HTTPMiddleware(nil)

// 自定义配置
mw := secure.HTTPMiddleware(&secure.Config{
    XFrameOptions:         "SAMEORIGIN",
    HSTSMaxAge:            63072000,        // 2 年
    HSTSIncludeSubdomains: true,
    HSTSPreload:           true,
    ContentSecurityPolicy: "default-src 'self'",
    IsDevelopment:         false,           // true 时跳过 HSTS（本地开发用）
})

// 注入 httpserver
srv := httpserver.New(mux,
    httpserver.WithMiddlewares(secure.HTTPMiddleware(nil)),
)
```

**默认设置的头部：**
- `X-Frame-Options: DENY`
- `X-Content-Type-Options: nosniff`
- `X-XSS-Protection: 1; mode=block`
- `Strict-Transport-Security: max-age=31536000; includeSubDomains`
- `Referrer-Policy: strict-origin-when-cross-origin`

## csrf — CSRF 防护

```go
import "github.com/Tsukikage7/servex/middleware/csrf"

// 使用默认配置（Double Submit Cookie 模式）
mw := csrf.HTTPMiddleware(nil)

// 自定义配置
mw := csrf.HTTPMiddleware(&csrf.Config{
    CookieName:   "_csrf",
    HeaderName:   "X-CSRF-Token",   // 前端通过此 header 回传 token
    FormField:    "csrf_token",      // 或表单字段
    CookieMaxAge: 12 * time.Hour,
    Secure:       true,
    SameSite:     http.SameSiteStrictMode,
    Skipper: func(r *http.Request) bool {
        return strings.HasPrefix(r.URL.Path, "/webhook") // 跳过 webhook 回调
    },
})

// 在 handler 中读取 token（用于渲染到页面或 JSON 响应）
func myHandler(w http.ResponseWriter, r *http.Request) {
    token := csrf.TokenFromContext(r.Context())
    json.NewEncoder(w).Encode(map[string]string{"csrf_token": token})
}
```

**工作流程：**
1. GET 请求 → 生成 token → 写入 `_csrf` cookie → 注入 context
2. POST/PUT/DELETE → 读取 `_csrf` cookie + `X-CSRF-Token` header → 恒定时间比较

## bodylimit — 请求体大小限制

```go
import "github.com/Tsukikage7/servex/middleware/bodylimit"

// 直接指定字节数（1 MB）
mw := bodylimit.HTTPMiddleware(1 << 20)

// 使用 ParseLimit 解析人类可读大小
limit, err := bodylimit.ParseLimit("10MB") // 支持 B/KB/MB/GB/TB
if err != nil {
    log.Fatal(err)
}
mw := bodylimit.HTTPMiddleware(limit)

// 注入 httpserver
srv := httpserver.New(mux,
    httpserver.WithMiddlewares(
        bodylimit.HTTPMiddleware(1 << 20), // API 路由限制 1 MB
    ),
)
```

**超限响应：** 返回 `413 Request Entity Too Large`

**实现：** Content-Length 快速检查 + `http.MaxBytesReader` 兜底，防止绕过

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

## signature — HMAC 请求签名

```go
// 服务端验签中间件
cfg := signature.DefaultConfig("shared-secret")
// 自定义
cfg = &signature.Config{
    Secret:          "shared-secret",
    HeaderName:      "X-Signature",    // 默认
    TimestampHeader: "X-Timestamp",    // 默认
    MaxAge:          5 * time.Minute,  // 防重放窗口
    Algorithm:       "sha256",         // "sha256" 或 "sha512"
}
handler = signature.HTTPMiddleware(cfg)(handler)
```

```go
// 客户端签名（自动设置 X-Timestamp + X-Signature header）
req, _ := http.NewRequest("POST", url, body)
_ = signature.SignRequest(req, "shared-secret")
// 或使用自定义配置
_ = signature.SignRequestWithConfig(req, cfg)
```

**签名算法：** `HMAC-SHA256(secret, timestamp + "." + body)`

**错误：** `ErrMissingSignature` / `ErrMissingTimestamp` / `ErrExpiredTimestamp` / `ErrInvalidSignature`（均返回 401）

**低级 API：**
- `signature.Sign(body, timestamp, secret)` — 计算签名
- `signature.Verify(body, timestamp, sig, secret)` — 常量时间比较验证

## trace — 链路追踪增强

```go
import "github.com/Tsukikage7/servex/middleware/trace"
```

统一 trace-id 在日志、响应头、下游调用中的传播，构建于 `middleware/requestid` 和 `observability/tracing` 之上。

```go
// HTTP 中间件
mw := trace.HTTPMiddleware(nil) // 使用默认配置

// 自定义配置
mw := trace.HTTPMiddleware(&trace.Config{
    TraceIDHeader:    "X-Trace-ID",    // 默认
    RequestIDHeader:  "X-Request-ID",  // 默认
    Logger:           log,             // 自动注入 trace_id 字段到日志
})

// 注入 httpserver（放在 requestid 之后）
srv := httpserver.New(mux,
    httpserver.WithMiddlewares(
        requestid.New(),
        trace.HTTPMiddleware(nil),
    ),
)
```

```go
// gRPC 拦截器
unaryInterceptor  := trace.GRPCUnaryInterceptor(nil)
streamInterceptor := trace.GRPCStreamInterceptor(nil)
```

```go
// 在 handler 中读取
traceID := trace.TraceIDFromContext(ctx)
reqID   := trace.RequestIDFromContext(ctx)

// 向下游传播（HTTP 客户端调用）
trace.InjectHTTPHeaders(ctx, req)

// 向下游传播（gRPC 客户端调用）
ctx = trace.InjectGRPCMetadata(ctx)
```

**默认行为：**
1. 从请求头（`X-Trace-ID`）提取 trace-id，不存在则生成 UUID
2. 优先从 `requestid` 中间件获取 request-id，其次从请求头，最后生成 UUID
3. 将 trace-id / request-id 写入响应头
4. 注入 logger context（后续 `log.Info(ctx, ...)` 自动携带 `trace_id` 字段）

**关键 API：**
- `trace.HTTPMiddleware(cfg)` — HTTP 中间件
- `trace.GRPCUnaryInterceptor(cfg)` / `GRPCStreamInterceptor(cfg)` — gRPC 拦截器
- `trace.TraceIDFromContext(ctx)` / `RequestIDFromContext(ctx)` — 读取 context
- `trace.InjectHTTPHeaders(ctx, req)` — 传播到下游 HTTP 请求
- `trace.InjectGRPCMetadata(ctx)` — 传播到下游 gRPC 调用
- `trace.DefaultConfig()` — 默认配置（TraceIDHeader: `X-Trace-ID`，RequestIDHeader: `X-Request-ID`）

---
name: observability
description: servex 可观测性模块专家。当用户使用 servex 的 observability/metrics（Prometheus）、observability/tracing（OpenTelemetry）或 observability/logger（结构化日志）时触发。
---

# servex 可观测性

## observability/metrics — Prometheus 指标

```go
// MustNewMetrics 初始化失败直接 panic（适合 main 函数）
m := metrics.MustNewMetrics(metrics.DefaultConfig("my-service"))

// NewMetrics 返回 error
m, err := metrics.NewMetrics(metrics.DefaultConfig("my-service"))
if err != nil { ... }

// HTTP 中间件（自动记录请求数、延迟、状态码）
mux.Handle("/metrics", promhttp.Handler())
srv := httpserver.New(mux,
    httpserver.WithMiddlewares(m.HTTPMiddleware()),
)
```

**关键选项：**
- `metrics.DefaultConfig(serviceName)` — 默认配置，注册 HTTP/gRPC 指标
- `m.HTTPMiddleware()` — `func(http.Handler) http.Handler`
- `m.GRPCUnaryInterceptor()` — gRPC 一元拦截器

## observability/tracing — OpenTelemetry 追踪

```go
// OTLP HTTP 导出（Jaeger、Grafana Tempo 等）
tracer, err := tracing.NewTracer(tracing.TracingConfig{
    ServiceName: "my-service",
    OTLP: &tracing.OTLPConfig{
        Endpoint: "http://localhost:4318", // OTLP HTTP 端口
    },
})
if err != nil { ... }
defer tracer.Shutdown(ctx)

// MustNewTracer 初始化失败直接 panic
tracer := tracing.MustNewTracer(tracing.TracingConfig{...})

// 与 httpserver 集成（自动注入 trace ID 到 context）
srv := httpserver.New(mux,
    httpserver.WithTrace("my-service"), // 快捷选项，内部使用默认 tracer
)
```

**与 logging 配合（组合顺序）：**

```
requestid → logging → tracing → metrics → ...
```

logging 在 tracing 之前：tracing 将 trace ID 写入 context，logging 可在后续请求处理中提取并输出。

## observability/logger — 结构化日志

```go
// 创建 logger（基于 zap）
log, err := logger.NewLogger(&logger.Config{
    Type:        logger.TypeZap,
    ServiceName: "my-service",
    Level:       logger.LevelInfo,     // debug/info/warn/error/fatal/panic
    Format:      logger.FormatJSON,    // json / console
    Output:      logger.OutputBoth,    // console / file / both
    LogDir:      "./logs",
    LevelSeparate: true,               // 按级别分文件
    RotationEnabled: true,
    RotationTime: logger.RotationDaily,
    MaxAge:      7,                    // 日志保留天数
    Compress:    true,
    EnableCaller: true,
    EnableStacktrace: false,
    TimeFormat:  logger.TimeFormatISO8601,
})
if err != nil { ... }
defer log.Close()

// MustNewLogger 失败时 panic
log := logger.MustNewLogger(&logger.Config{...})

// 基础日志
log.Info("服务启动")
log.Errorf("请求失败: %v", err)

// 结构化字段
log.With(
    logger.Field{Key: "user_id", Value: "u-1"},
    logger.Field{Key: "latency_ms", Value: 42},
).Info("请求完成")

// 注入 context（自动提取 traceId/spanId）
log.WithContext(ctx).Info("带链路追踪的日志")
```

**Logger 接口方法：**
- 级别方法：`Debug`/`Info`/`Warn`/`Error`/`Fatal`/`Panic`（及 `f` 格式化版本）
- `With(fields...) Logger` — 附加结构化字段
- `WithContext(ctx) Logger` — 注入 context（自动提取 traceId）
- `Sync() error` / `Close() error` — 刷新/关闭

**辅助函数：**
- `logger.ContextWithTraceID(ctx, traceID)` — 注入 traceId 到 context
- `logger.ContextWithSpanID(ctx, spanID)` — 注入 spanId 到 context

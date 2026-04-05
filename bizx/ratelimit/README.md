# bizx/ratelimit — 业务配额管理

提供按用户/租户维度的配额管理，区别于 `middleware/ratelimit` 的请求级限流，本包关注**业务配额语义**（已用量、剩余量、重置时间）。

## 实现

| 构造函数 | 说明 |
|----------|------|
| `NewMemoryQuotaManager()` | 内存实现，适合测试 |
| `NewRedisQuotaManager(client, opts...)` | Redis 实现，分布式场景 |

## 接口

```go
type QuotaManager interface {
    Check(ctx, quota Quota) (*Usage, error)          // 检查配额（不消耗）
    Consume(ctx, quota Quota, n int64) (*Usage, error) // 消耗配额
    Reset(ctx, key string) error                     // 重置配额
    GetUsage(ctx, quota Quota) (*Usage, error)        // 获取使用量
}
```

## 数据结构

```go
type Quota struct {
    Key    string        // 配额键（如 "user:123"、"tenant:abc"）
    Limit  int64         // 配额上限
    Window time.Duration // 配额窗口（如 24h、720h=30天）
}

type Usage struct {
    Used      int64     // 已用量
    Remaining int64     // 剩余量
    Limit     int64     // 上限
    ResetsAt  time.Time // 重置时间
}
```

## 选项

| 选项 | 说明 |
|------|------|
| `WithKeyPrefix(prefix)` | Redis key 前缀 |

## 快速上手

```go
mgr := ratelimit.NewRedisQuotaManager(redisClient,
    ratelimit.WithKeyPrefix("myapp:"))

quota := ratelimit.Quota{
    Key:    "user:" + userID,
    Limit:  1000,
    Window: 24 * time.Hour, // 每天 1000 次
}

// 消耗配额
usage, err := mgr.Consume(ctx, quota, 1)
if errors.Is(err, ratelimit.ErrQuotaExceeded) {
    // 返回 429，附带剩余配额信息
    return fmt.Errorf("配额已耗尽，将于 %v 重置", usage.ResetsAt)
}

// 在响应头中返回配额信息
w.Header().Set("X-RateLimit-Limit", strconv.FormatInt(usage.Limit, 10))
w.Header().Set("X-RateLimit-Remaining", strconv.FormatInt(usage.Remaining, 10))
```

## 错误

| 错误 | 说明 |
|------|------|
| `ErrQuotaExceeded` | 配额已用尽，同时返回当前 Usage |

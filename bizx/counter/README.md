# bizx/counter — 分布式计数器

提供业务级计数能力，区别于 `storage/redis` 的原子操作，本包额外支持**滑动窗口统计**和**批量获取**。

## 实现

| 构造函数 | 说明 |
|----------|------|
| `NewMemoryCounter(opts...)` | 内存实现，适合测试或单进程场景 |
| `NewRedisCounter(client, opts...)` | Redis 实现，适合分布式场景 |

## 接口

```go
type Counter interface {
    Incr(ctx, key string, delta int64) (int64, error)      // 增加计数
    Get(ctx, key string) (int64, error)                    // 获取当前值
    Reset(ctx, key string) error                           // 重置
    IncrWindow(ctx, key string, window Duration) (int64, error) // 滑动窗口计数
    GetWindow(ctx, key string, window Duration) (int64, error)  // 获取窗口内计数
    MGet(ctx, keys ...string) (map[string]int64, error)    // 批量获取
}
```

## 选项

| 选项 | 说明 |
|------|------|
| `WithPrefix(prefix)` | 设置 Redis key 前缀 |

## 快速上手

```go
// 分布式场景使用 Redis 实现
c := counter.NewRedisCounter(redisClient, counter.WithPrefix("myapp:"))

// 简单计数（如登录次数）
val, _ := c.Incr(ctx, "login:user:123", 1)

// 滑动窗口（如最近 5 分钟 API 调用次数）
count, _ := c.IncrWindow(ctx, "api:user:123", 5*time.Minute)

// 批量获取多个计数器
counts, _ := c.MGet(ctx, "login:user:1", "login:user:2", "login:user:3")
```

## Redis 实现细节

- 简单计数：使用 Redis `INCRBY`
- 滑动窗口：使用 Redis Sorted Set（ZSet），以 Unix 纳秒时间戳为 score，自动清理过期成员
- 批量获取：使用 Redis `MGET` Pipeline

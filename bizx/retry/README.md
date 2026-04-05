# bizx/retry — 异步持久化重试

将失败任务持久化到存储（GORM/内存），后台调度器定期轮询并按**指数退避**策略重试，超出最大次数后进入**死信**状态。

## 实现

| 存储构造函数 | 说明 |
|-------------|------|
| `NewMemoryStore()` | 内存存储，适合测试 |
| `NewGORMStore(db)` | GORM 存储（需运行 `AutoMigrate`） |

## 数据结构

```go
type Task struct {
    ID          string          // 任务 ID
    Name        string          // 任务名（对应注册的 Handler）
    Payload     json.RawMessage // 任务载荷
    MaxRetries  int             // 最大重试次数
    Retried     int             // 已重试次数
    NextRetryAt time.Time       // 下次重试时间
    Status      Status          // pending/running/done/dead
    LastError   string          // 最后一次错误信息
}
```

## 接口

```go
type Scheduler interface {
    Submit(ctx, name string, payload any, opts ...TaskOption) (string, error)
    Register(name string, handler Handler)
    Start(ctx) error
    Stop(ctx) error
}

type Handler func(ctx context.Context, payload json.RawMessage) error
```

## 调度器选项

| 选项 | 默认值 | 说明 |
|------|--------|------|
| `WithPollInterval(d)` | 10s | 轮询间隔 |
| `WithConcurrency(n)` | 5 | 并发执行数 |

## 任务选项

| 选项 | 默认值 | 说明 |
|------|--------|------|
| `WithMaxRetries(n)` | 5 | 最大重试次数 |
| `WithInitialDelay(d)` | 1m | 初始延迟 |
| `WithBackoffMultiplier(m)` | 2.0 | 退避倍数 |

## 快速上手

```go
store := retry.NewGORMStore(db)
store.AutoMigrate(ctx)

scheduler := retry.NewScheduler(store,
    retry.WithPollInterval(30*time.Second),
    retry.WithConcurrency(10),
)

// 注册处理器
scheduler.Register("send_email", func(ctx context.Context, payload json.RawMessage) error {
    var req EmailRequest
    json.Unmarshal(payload, &req)
    return emailService.Send(ctx, req)
})

scheduler.Start(ctx)
defer scheduler.Stop(ctx)

// 提交任务（失败时自动重试）
taskID, _ := scheduler.Submit(ctx, "send_email",
    EmailRequest{To: "user@example.com", Subject: "Welcome"},
    retry.WithMaxRetries(3),
)
```

## 重试间隔（指数退避）

第 1 次重试后等待 2^0 = 1 分钟，第 2 次等待 2^1 = 2 分钟，第 3 次等待 2^2 = 4 分钟，依此类推。

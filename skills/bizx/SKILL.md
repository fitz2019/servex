---
name: bizx
description: servex 业务组件。当用户需要排行榜、计数器、序号生成、状态机、审计日志、特性开关、验证码、业务锁、业务限流、游标分页、异步重试、进程内事件总线时触发。
---

# bizx 业务组件使用指南

`bizx/` 是 servex 的业务组件集合，每个子包解决一个常见的业务场景需求，均提供内存实现（测试）和 Redis/GORM 实现（生产）。

---

## bizx/counter — 分布式计数器

**包路径：** `github.com/Tsukikage7/servex/bizx/counter`

**何时使用：** 需要精确计数（登录次数、API 调用量）或滑动窗口统计时。

### 接口

```go
type Counter interface {
    Incr(ctx, key string, delta int64) (int64, error)
    Get(ctx, key string) (int64, error)
    Reset(ctx, key string) error
    IncrWindow(ctx, key string, window time.Duration) (int64, error)
    GetWindow(ctx, key string, window time.Duration) (int64, error)
    MGet(ctx, keys ...string) (map[string]int64, error)
}
```

### 构造函数

- `NewMemoryCounter(opts...)` — 内存实现（测试/单进程）
- `NewRedisCounter(client, opts...)` — Redis 实现（分布式）
- `WithPrefix(prefix)` — 设置 key 前缀

### 示例

```go
c := counter.NewRedisCounter(redisClient, counter.WithPrefix("myapp:"))

// 精确计数
val, _ := c.Incr(ctx, "login:user:123", 1)

// 滑动窗口（最近 5 分钟 API 调用次数）
count, _ := c.IncrWindow(ctx, "api:user:123", 5*time.Minute)

// 批量获取
counts, _ := c.MGet(ctx, "login:user:1", "login:user:2")
```

---

## bizx/leaderboard — 排行榜

**包路径：** `github.com/Tsukikage7/servex/bizx/leaderboard`

**何时使用：** 需要 Top N、分页排行、按分数排序的场景（游戏积分榜、销售排名、活跃度榜单）。

### 接口

```go
type Leaderboard interface {
    AddScore(ctx, member string, score float64) error
    IncrScore(ctx, member string, delta float64) (float64, error)
    GetRank(ctx, member string) (*Entry, error)
    GetScore(ctx, member string) (float64, error)
    TopN(ctx, n int) ([]Entry, error)
    GetPage(ctx, offset, limit int) (*Page, error)
    Remove(ctx, members ...string) error
    Count(ctx) (int64, error)
    Reset(ctx) error
}
```

### 构造函数

- `NewMemoryLeaderboard(name, opts...)` — 内存实现
- `NewRedisLeaderboard(client, name, opts...)` — Redis Sorted Set 实现
- `WithOrder(Descending|Ascending)` — 排序方式（默认降序）
- `WithPrefix(prefix)` — key 前缀

### 示例

```go
lb := leaderboard.NewRedisLeaderboard(redisClient, "weekly_score")

lb.AddScore(ctx, "player1", 1500)
lb.IncrScore(ctx, "player1", 200)

top10, _ := lb.TopN(ctx, 10)
for _, e := range top10 {
    fmt.Printf("Rank %d: %s %.0f\n", e.Rank, e.Member, e.Score)
}

page, _ := lb.GetPage(ctx, 20, 20) // 第 2 页，每页 20 条
```

---

## bizx/sequence — 业务序号生成器

**包路径：** `github.com/Tsukikage7/servex/bizx/sequence`

**何时使用：** 需要生成有意义的业务编号（订单号、单据号），如 `ORD-20260405-0001`。

### 接口

```go
type Sequence interface {
    Next(ctx) (string, error)
    Current(ctx) (string, error)
    Reset(ctx) error
}
```

### Config

```go
type Config struct {
    Name       string // 序列名
    Prefix     string // 前缀（如 "ORD-"）
    DateFormat string // 日期格式（如 "20060102"）
    Padding    int    // 补零位数，默认 4
    Step       int64  // 步长，默认 1
    ResetDaily bool   // 每日重置
}
```

### 构造函数

- `New(cfg, store)` — 创建序号生成器
- `NewMemoryStore()` — 内存存储
- `NewRedisStore(client)` — Redis 存储

### 示例

```go
store := sequence.NewRedisStore(redisClient)
seq := sequence.New(&sequence.Config{
    Name:       "order",
    Prefix:     "ORD-",
    DateFormat: "20060102",
    Padding:    4,
    ResetDaily: true,
}, store)

id, _ := seq.Next(ctx) // "ORD-20260405-0001"
id, _ = seq.Next(ctx)  // "ORD-20260405-0002"
```

---

## bizx/locking — 业务锁

**包路径：** `github.com/Tsukikage7/servex/bizx/locking`

**何时使用：** 需要可重入锁、读写锁或带续期的业务锁，区别于 `storage/lock` 的基础分布式锁。

### 接口

```go
type Lock interface {
    Lock(ctx) error
    Unlock(ctx) error
    Extend(ctx, ttl Duration) error
}
type ReentrantLock interface { Lock; LockCount() int }
type RWLock interface { RLock(ctx) error; RUnlock(ctx) error; Lock(ctx) error; Unlock(ctx) error }
```

### 构造函数

- `NewLock(locker, key, opts...)` — 普通分布式锁
- `NewReentrantLock(locker, key, opts...)` — 可重入锁
- `NewRWLock(locker, key, opts...)` — 读写锁
- `WithTTL(d)` 默认 30s；`WithRetryInterval(d)` 默认 100ms；`WithRetryTimeout(d)` 默认 10s

### 示例

```go
locker, _ := storagelock.NewLocker(redisClient)

// 推荐：WithLock 辅助函数
l := locking.NewLock(locker, "order:"+orderID)
err := locking.WithLock(ctx, l, func(ctx context.Context) error {
    return processOrder(ctx, orderID)
})

// 读写锁
rwl := locking.NewRWLock(locker, "config")
locking.WithRLock(ctx, rwl, func(ctx context.Context) error {
    return readConfig(ctx)
})
```

---

## bizx/ratelimit — 业务配额管理

**包路径：** `github.com/Tsukikage7/servex/bizx/ratelimit`

**何时使用：** 需要按用户/租户进行配额控制，并能查看已用量、剩余量、重置时间（区别于 `middleware/ratelimit` 的无状态限流）。

### 接口

```go
type QuotaManager interface {
    Check(ctx, quota Quota) (*Usage, error)
    Consume(ctx, quota Quota, n int64) (*Usage, error)
    Reset(ctx, key string) error
    GetUsage(ctx, quota Quota) (*Usage, error)
}
```

### 数据结构

```go
type Quota struct { Key string; Limit int64; Window time.Duration }
type Usage struct { Used, Remaining, Limit int64; ResetsAt time.Time }
```

### 构造函数

- `NewMemoryQuotaManager()` — 内存实现
- `NewRedisQuotaManager(client, opts...)` — Redis 实现
- `WithKeyPrefix(prefix)` — key 前缀

### 示例

```go
mgr := ratelimit.NewRedisQuotaManager(redisClient)
quota := ratelimit.Quota{Key: "user:" + userID, Limit: 1000, Window: 24 * time.Hour}

usage, err := mgr.Consume(ctx, quota, 1)
if errors.Is(err, ratelimit.ErrQuotaExceeded) {
    return fmt.Errorf("配额耗尽，%v 后重置", time.Until(usage.ResetsAt))
}
```

---

## bizx/statemachine — 有限状态机

**包路径：** `github.com/Tsukikage7/servex/bizx/statemachine`

**何时使用：** 订单流程、工单流转、审批流等需要明确定义状态转换的场景。

### 核心类型

```go
type Transition struct {
    From, To State
    Event    Event
    Guard    func(ctx context.Context, data any) bool
    Action   func(ctx context.Context, data any) error
}
```

### API

- `New(initial, transitions)` — 创建状态机
- `Fire(ctx, event, data)` — 触发事件
- `Current()` — 当前状态
- `Can(event)` — 是否可触发
- `AvailableEvents()` — 可用事件列表
- `OnEnter(state, fn)` / `OnLeave(state, fn)` / `OnTransition(fn)` — 回调注册

### 示例

```go
sm := statemachine.New("pending", []statemachine.Transition{
    {From: "pending", Event: "pay",     To: "paid"},
    {From: "paid",    Event: "ship",    To: "shipped"},
    {From: "shipped", Event: "deliver", To: "delivered"},
    {From: "pending", Event: "cancel",  To: "cancelled"},
})

sm.OnEnter("paid", func(ctx context.Context, data any) {
    sendConfirmEmail(ctx, data.(*Order))
})

sm.Fire(ctx, "pay", order) // pending → paid
sm.Current()               // "paid"
sm.Can("ship")             // true
```

---

## bizx/pagination — 游标分页

**包路径：** `github.com/Tsukikage7/servex/bizx/pagination`

**何时使用：** 大数据量列表、实时数据流，避免 OFFSET 分页的性能问题。

### 核心函数

- `EncodeCursor(values...)` — 编码游标（base64url+JSON）
- `DecodeCursor(cursor)` — 解码游标
- `GORMPaginate(db, req, orderField)` — GORM 集成辅助

### 数据结构

```go
type CursorRequest struct { Cursor string; Limit int; Direction Direction }
type CursorResponse[T any] struct { Items []T; NextCursor, PrevCursor string; HasMore bool }
```

### 示例

```go
req := (&pagination.CursorRequest{
    Cursor: r.URL.Query().Get("cursor"),
    Limit:  20,
}).Apply()

var posts []Post
pagination.GORMPaginate(db, req, "id").Find(&posts)

hasMore := len(posts) > req.Limit
if hasMore { posts = posts[:req.Limit] }

var nextCursor string
if hasMore && len(posts) > 0 {
    nextCursor = pagination.EncodeCursor(posts[len(posts)-1].ID)
}
```

---

## bizx/audit — 审计日志

**包路径：** `github.com/Tsukikage7/servex/bizx/audit`

**何时使用：** 需要记录操作历史（谁/对什么/做了什么/改了哪些字段）的场景。

### 接口

```go
type Logger interface {
    Log(ctx, entry *Entry) error
    Query(ctx, filter *Filter) ([]Entry, error)
}
```

### 核心类型

```go
type Entry struct {
    Actor, Action, Resource, ResourceID string
    Changes  map[string]Change  // {field: {From, To}}
    Metadata map[string]any
    IP, UserAgent string
    CreatedAt time.Time
}
```

### 构造函数

- `NewLogger(store, opts...)` — 创建记录器
- `WithAsync(bufferSize)` — 异步写入
- `NewMemoryStore()` / `NewGORMStore(db)` — 存储实现
- `HTTPMiddleware(logger, opts...)` — HTTP 中间件

### 示例

```go
store := audit.NewGORMStore(db)
store.AutoMigrate(ctx)
auditLog := audit.NewLogger(store, audit.WithAsync(1024))

auditLog.Log(ctx, &audit.Entry{
    Actor: userID, Action: "UPDATE", Resource: "order", ResourceID: orderID,
    Changes: map[string]audit.Change{
        "status": {From: "pending", To: "paid"},
    },
})

entries, _ := auditLog.Query(ctx, &audit.Filter{Actor: userID, Limit: 50})
```

---

## bizx/feature — 特性开关

**包路径：** `github.com/Tsukikage7/servex/bizx/feature`

**何时使用：** 需要灰度发布、A/B 测试、白名单放量等功能。

### 接口

```go
type Manager interface {
    IsEnabled(ctx, name string, opts ...EvalOption) bool
    GetFlag(ctx, name string) (*Flag, error)
    SetFlag(ctx, flag *Flag) error
    DeleteFlag(ctx, name string) error
    ListFlags(ctx) ([]*Flag, error)
}
```

### Flag 结构

```go
type Flag struct {
    Name       string
    Enabled    bool         // 全局开关
    Percentage int          // 百分比放量（0-100）
    Users      []string     // 用户白名单
    Groups     []string     // 分组白名单
    Metadata   map[string]any
}
```

### 构造函数

- `NewManager(store)` — 创建管理器
- `NewMemoryStore()` — 内存存储
- `NewRedisStore(client, opts...)` — Redis 存储
- `WithUser(id)` / `WithGroup(g)` / `WithAttributes(attrs)` — 评估选项

### 示例

```go
mgr := feature.NewManager(feature.NewRedisStore(redisClient))

mgr.SetFlag(ctx, &feature.Flag{
    Name: "new_ui", Enabled: true, Percentage: 20,
    Users: []string{"beta_user_1"},
})

if mgr.IsEnabled(ctx, "new_ui", feature.WithUser(userID)) {
    renderNewUI(w)
}
```

---

## bizx/retry — 异步持久化重试

**包路径：** `github.com/Tsukikage7/servex/bizx/retry`

**何时使用：** 需要将失败操作持久化后异步重试（发送通知、调用第三方 API 等），防止数据丢失。

### 接口

```go
type Scheduler interface {
    Submit(ctx, name string, payload any, opts ...TaskOption) (string, error)
    Register(name string, handler Handler)
    Start(ctx) error
    Stop(ctx) error
}
type Handler func(ctx context.Context, payload json.RawMessage) error
```

### 构造函数

- `NewScheduler(store, opts...)` — 创建调度器
- `WithPollInterval(d)` 默认 10s；`WithConcurrency(n)` 默认 5
- `NewMemoryStore()` / `NewGORMStore(db)` — 存储实现
- `WithMaxRetries(n)` 默认 5；`WithInitialDelay(d)` 默认 1m；`WithBackoffMultiplier(m)` 默认 2.0

### 示例

```go
store := retry.NewGORMStore(db)
store.AutoMigrate(ctx)

s := retry.NewScheduler(store, retry.WithConcurrency(10))
s.Register("send_sms", func(ctx context.Context, payload json.RawMessage) error {
    var req SMSRequest
    json.Unmarshal(payload, &req)
    return smsClient.Send(ctx, req)
})
s.Start(ctx)
defer s.Stop(ctx)

s.Submit(ctx, "send_sms", SMSRequest{Phone: "138xxxx", Text: "验证码：1234"},
    retry.WithMaxRetries(3))
```

---

## bizx/event — 进程内事件总线

**包路径：** `github.com/Tsukikage7/servex/bizx/event`

**何时使用：** 模块间解耦，进程内事件驱动，不需要跨进程（跨进程请用 `messaging/pubsub`）。

### 接口

```go
type Bus interface {
    Publish(ctx, name string, payload any) error
    Subscribe(pattern string, handler Handler, opts ...SubOption)
    Unsubscribe(pattern string)
    Close() error
}
type Handler func(ctx context.Context, evt *Event) error
```

### 通配符

- `*` — 匹配所有事件
- `user.*` — 匹配 `user.created`、`user.deleted`（单层）
- `order.paid` — 精确匹配

### 构造函数

- `New(opts...)` — 创建事件总线
- `WithBufferSize(n)` 默认 1024；`WithErrorHandler(fn)` — 异步错误处理
- `WithPriority(n)` — 订阅优先级（越小越先）；`WithAsync(true)` — 异步执行

### 示例

```go
bus := event.New(event.WithBufferSize(2048))
defer bus.Close()

bus.Subscribe("user.*", func(ctx context.Context, evt *event.Event) error {
    log.Printf("user event: %s", evt.Name)
    return nil
})
bus.Subscribe("order.paid", sendEmailHandler, event.WithAsync(true))

bus.Publish(ctx, "user.created", UserCreatedEvent{UserID: "123"})
bus.Publish(ctx, "order.paid", OrderPaidEvent{OrderID: "456"})
```

---

## bizx/captcha — 验证码管理

**包路径：** `github.com/Tsukikage7/servex/bizx/captcha`

**何时使用：** 短信/邮件验证码场景，需要防暴力破解和防刷。

### 接口

```go
type Manager interface {
    Generate(ctx, key string) (*Code, error)
    Verify(ctx, key, code string) error
    Invalidate(ctx, key string) error
}
```

### 选项

| 选项 | 默认值 | 说明 |
|------|--------|------|
| `WithLength(n)` | 6 | 验证码长度 |
| `WithExpiration(d)` | 5m | 过期时间 |
| `WithMaxAttempts(n)` | 5 | 最大验证次数 |
| `WithCooldown(d)` | 60s | 发送冷却（防刷） |
| `WithAlphabet(s)` | `"0123456789"` | 字符集 |

### 构造函数

- `NewManager(store, opts...)` — 创建管理器
- `NewMemoryStore()` — 内存存储
- `NewRedisStore(client)` — Redis 存储

### 示例

```go
mgr := captcha.NewManager(
    captcha.NewRedisStore(redisClient),
    captcha.WithExpiration(10*time.Minute),
    captcha.WithCooldown(60*time.Second),
)

code, err := mgr.Generate(ctx, phone) // 生成，然后通过短信发送 code.Code
if errors.Is(err, captcha.ErrCooldown) { return errors.New("请稍后再试") }

err = mgr.Verify(ctx, phone, inputCode)
switch {
case errors.Is(err, captcha.ErrCodeExpired):    return errors.New("验证码已过期")
case errors.Is(err, captcha.ErrCodeInvalid):    return errors.New("验证码错误")
case errors.Is(err, captcha.ErrTooManyAttempts): return errors.New("尝试次数过多")
}
```

---

## 选择指南

| 需求 | 推荐模块 |
|------|----------|
| 记录某用户今天登录了多少次 | `bizx/counter` |
| 最近 5 分钟内 API 调用次数 | `bizx/counter` (IncrWindow) |
| 游戏积分榜/销售排名 | `bizx/leaderboard` |
| 生成订单号 ORD-20260405-0001 | `bizx/sequence` |
| 防止并发修改同一资源 | `bizx/locking` |
| 用户每天最多调用 API 1000 次 | `bizx/ratelimit` |
| 订单/工单流程状态管理 | `bizx/statemachine` |
| 无限滚动/大数据量分页 | `bizx/pagination` |
| 记录谁改了什么字段 | `bizx/audit` |
| 灰度发布/A/B 测试 | `bizx/feature` |
| 第三方 API 调用失败后自动重试 | `bizx/retry` |
| 模块间解耦事件通知 | `bizx/event` |
| 短信/邮件验证码 | `bizx/captcha` |

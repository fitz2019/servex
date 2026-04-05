# bizx/audit — 审计日志

记录**谁**对**什么**做了**什么**，支持字段变更追踪（before/after diff）、多存储后端、异步写入和 HTTP 中间件。

## 实现

| 存储构造函数 | 说明 |
|-------------|------|
| `NewMemoryStore()` | 内存存储，适合测试 |
| `NewGORMStore(db)` | GORM 存储（需运行 `AutoMigrate`） |

## 数据结构

```go
type Entry struct {
    ID         string            // 审计记录 ID
    Actor      string            // 操作者（如 userID）
    Action     string            // 操作（如 "UPDATE"、"DELETE"）
    Resource   string            // 资源类型（如 "order"）
    ResourceID string            // 资源 ID
    Changes    map[string]Change // 字段变更 {field: {From, To}}
    Metadata   map[string]any    // 附加元数据
    IP         string
    UserAgent  string
    CreatedAt  time.Time
}

type Change struct {
    From any
    To   any
}
```

## 接口

```go
type Logger interface {
    Log(ctx, entry *Entry) error
    Query(ctx, filter *Filter) ([]Entry, error)
}
```

## 创建记录器

```go
store := audit.NewGORMStore(db)
store.AutoMigrate(ctx) // 建表

// 同步写入
auditLog := audit.NewLogger(store)

// 异步写入（高吞吐场景，buffSize = 1024）
auditLog := audit.NewLogger(store, audit.WithAsync(1024))
```

## 快速上手

```go
// 手动记录
auditLog.Log(ctx, &audit.Entry{
    Actor:      currentUserID,
    Action:     "UPDATE",
    Resource:   "order",
    ResourceID: orderID,
    Changes: map[string]audit.Change{
        "status": {From: "pending", To: "paid"},
        "amount": {From: 100.0, To: 99.0},
    },
})

// 查询某用户的操作记录
entries, _ := auditLog.Query(ctx, &audit.Filter{
    Actor:    userID,
    Resource: "order",
    From:     time.Now().AddDate(0, -1, 0),
    Limit:    50,
})

// HTTP 中间件（自动记录所有请求）
http.Handle("/api/", audit.HTTPMiddleware(auditLog,
    audit.WithActorExtractor(func(r *http.Request) string {
        return r.Header.Get("X-User-ID")
    }),
)(apiHandler))
```

# bizx/event — 进程内事件总线

轻量进程内事件总线，支持**通配符订阅**、**优先级**、**同步/异步执行**。适合模块间解耦，不依赖外部消息队列。

> 如需跨进程事件，请使用 `messaging/pubsub`（Kafka/RabbitMQ/Redis Streams）。

## 接口

```go
type Bus interface {
    Publish(ctx, name string, payload any) error
    Subscribe(pattern string, handler Handler, opts ...SubOption)
    Unsubscribe(pattern string)
    Close() error
}

type Handler func(ctx context.Context, evt *Event) error

type Event struct {
    Name      string
    Payload   any
    Timestamp time.Time
}
```

## 通配符规则

| 模式 | 匹配示例 | 不匹配 |
|------|----------|--------|
| `*` | 所有事件 | - |
| `user.*` | `user.created`、`user.deleted` | `user.role.changed`（多层不匹配） |
| `order.paid` | `order.paid`（精确匹配） | `order.payment` |

## 总线选项

| 选项 | 默认值 | 说明 |
|------|--------|------|
| `WithBufferSize(n)` | 1024 | 异步队列容量 |
| `WithErrorHandler(fn)` | 忽略 | 异步处理器错误回调 |

## 订阅选项

| 选项 | 默认值 | 说明 |
|------|--------|------|
| `WithPriority(n)` | 0 | 优先级（数字越小越先执行） |
| `WithAsync(true)` | false | 异步执行（不阻塞 Publish） |

## 快速上手

```go
bus := event.New(
    event.WithBufferSize(2048),
    event.WithErrorHandler(func(err error) { log.Error(err) }),
)
defer bus.Close()

// 订阅用户相关事件（通配符）
bus.Subscribe("user.*", func(ctx context.Context, evt *event.Event) error {
    log.Printf("user event: %s payload: %v", evt.Name, evt.Payload)
    return nil
})

// 高优先级同步处理（如审计）
bus.Subscribe("order.paid", auditHandler, event.WithPriority(-10))

// 低优先级异步处理（如发邮件，不影响主流程）
bus.Subscribe("order.paid", sendEmailHandler, event.WithAsync(true))

// 发布事件
bus.Publish(ctx, "user.created", UserCreatedEvent{UserID: "123"})
bus.Publish(ctx, "order.paid", OrderPaidEvent{OrderID: "456", Amount: 99.9})
```

## 错误

| 错误 | 说明 |
|------|------|
| `ErrBusClosed` | 向已关闭的总线发布事件 |

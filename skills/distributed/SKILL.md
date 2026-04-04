---
name: distributed
description: servex 分布式模式专家。当用户使用 servex 的 cqrs（命令查询分离）、outbox（事务消息）、domain（领域事件总线）、saga（分布式事务编排）、eventsourcing（事件溯源）时触发。
---

# servex 分布式模式

## cqrs — 命令查询分离

```go
// CommandHandler 接口：Handle(ctx, command) (result, error)
type CreateOrderHandler struct{}

func (h *CreateOrderHandler) Handle(ctx context.Context, cmd CreateOrderCommand) (*Order, error) {
    // 执行业务逻辑
    return &Order{ID: uuid.New()}, nil
}

// 包装中间件链（日志 → 追踪 → metrics → handler）
handler := cqrs.ChainCommand[CreateOrderCommand, *Order](
    &CreateOrderHandler{},
    cqrsmw.CommandLogging[CreateOrderCommand, *Order](log),
    cqrsmw.CommandTracing[CreateOrderCommand, *Order]("create-order"),
)

// 执行命令
order, err := cqrs.ApplyCommand(ctx, handler, CreateOrderCommand{UserID: "u-1"})
```

**Query 模式（同结构）：**

```go
// QueryHandler：Handle(ctx, query) (result, error)
handler := cqrs.ChainQuery[GetOrderQuery, *Order](
    &GetOrderHandler{},
    cqrsmw.QueryLogging[GetOrderQuery, *Order](log),
)

order, err := cqrs.ApplyQueryHandler(ctx, handler, GetOrderQuery{OrderID: "o-1"})
```

**注意：** servex CQRS 无 Bus 注册表，直接持有 handler 引用，类型安全。

## outbox — Outbox 事务消息

```go
// 创建 Store（需要 GORM DB）
store := outbox.NewGORMStore(gormDB)
if err := store.AutoMigrate(); err != nil { ... } // 建表

// 在事务中保存消息（消息与业务数据同事务）
err := store.WithTx(ctx, func(ctx context.Context) error {
    tx := outbox.ExtractTx(ctx) // 获取事务内的 *gorm.DB
    // ... 执行业务写入 ...
    msg := outbox.NewOutboxMessage("order.created", payload)
    return store.Save(ctx, msg) // 与业务写入同事务
})

// 启动 Relay（后台轮询，将 Outbox 消息发送到消息队列）
relay := outbox.NewRelay(store, publisher)
if err := relay.Start(ctx); err != nil { ... }
defer relay.Stop()
```

**InjectTx / ExtractTx 模式：**

```go
// 注入事务到 context（供下游使用）
ctx = outbox.InjectTx(ctx, tx)

// 从 context 提取事务
tx := outbox.ExtractTx(ctx)
```

## domain — 领域事件总线

```go
// 同步事件总线
bus := domain.NewEventBus()

// 订阅特定事件
bus.Subscribe("order.created", domain.EventHandler(func(ctx context.Context, e domain.Event) error {
    fmt.Println("订单已创建:", e.AggregateID())
    return nil
}))

// 订阅所有事件
bus.SubscribeAll(domain.EventHandler(func(ctx context.Context, e domain.Event) error {
    fmt.Println("事件:", e.EventType())
    return nil
}))

// 发布事件（同步，等待所有 handler 完成）
if err := bus.Publish(ctx, orderCreatedEvent); err != nil { ... }

// 异步事件总线（handler 在 goroutine 中并发执行）
asyncBus := domain.NewAsyncEventBus()

// Outbox + Domain 桥接（将领域事件通过 Outbox 持久化后发布）
publisher := outbox.NewOutboxPublisher(store, domain.NewJSONEventConverter())
// publisher 实现 domain.EventBus 接口，可替换 asyncBus
```

**JSONEventConverter：**

```go
// 将 domain.Event 序列化为 JSON 存入 Outbox
converter := domain.NewJSONEventConverter()
```

## domain/saga — Saga 分布式事务编排

```go
// 定义步骤函数
reserveInventory := func(ctx context.Context, data *saga.Data) error {
    orderID := data.GetString("order_id")
    // 执行库存预留
    data.Set("reservation_id", "RES-123") // 步骤间传递数据
    return nil
}
compensateInventory := func(ctx context.Context, data *saga.Data) error {
    reservationID := data.GetString("reservation_id")
    // 回滚库存预留（补偿应幂等）
    return cancelReservation(ctx, reservationID)
}

// 使用 Builder 模式构建 Saga
s := saga.New("create-order").
    Step("reserve-inventory", reserveInventory, compensateInventory).
    Step("charge-payment", chargePayment, refundPayment).
    Step("send-notification", sendNotification, nil). // nil 表示无需补偿
    Options(
        saga.WithLogger(log),
        saga.WithTimeout(30 * time.Second),
        saga.WithRetry(2, time.Second),       // 步骤失败重试 2 次
        saga.WithStore(redisStore),            // 持久化状态（生产建议）
        saga.WithStepHooks(
            func(name string) { fmt.Println("开始:", name) },
            func(name string, err error) { fmt.Println("结束:", name, err) },
        ),
    ).
    Build()

// 执行（失败时自动逆序补偿）
err := s.Execute(ctx)

// 带初始数据执行
data := saga.NewData()
data.Set("order_id", "ORD-001")
err := s.ExecuteWithData(ctx, data)
```

**关键类型：**
- `saga.New(name) *Builder` — 创建构建器
- `builder.Step(name, action, compensate)` — 添加步骤（补偿可为 nil）
- `builder.Build() *Saga` — 构建 Saga 实例
- `saga.StepFunc` — `func(ctx, *Data) error`，步骤执行函数
- `saga.CompensateFunc` — `func(ctx, *Data) error`，补偿函数
- `saga.Data` — 步骤间共享数据（`Set`, `Get`, `GetString`, `GetInt`）
- 选项：`WithStore`, `WithLogger`, `WithTimeout`, `WithRetry(count, delay)`, `WithStepHooks`
- 状态流转：Pending → Running → Completed / Compensating → Compensated / CompensateFailed

## domain/eventsourcing — 事件溯源

```go
// 1. 定义领域聚合，嵌入 BaseAggregate
type Order struct {
    eventsourcing.BaseAggregate
    Status    string
    TotalAmt  int64
}

func NewOrder(id string) *Order {
    return &Order{BaseAggregate: eventsourcing.NewBaseAggregate(id, "Order")}
}

// 实现 Aggregate 接口：应用事件到聚合状态
func (o *Order) ApplyEvent(event eventsourcing.Event) error {
    switch event.EventType {
    case "OrderCreated":
        var data struct{ TotalAmt int64 }
        json.Unmarshal(event.Data, &data)
        o.Status = "created"
        o.TotalAmt = data.TotalAmt
    case "OrderShipped":
        o.Status = "shipped"
    }
    return nil
}

// 2. 发起事件（自动序列化 data，自增版本号，调用 ApplyEvent）
order := NewOrder("ORD-001")
err := order.RaiseEvent(order.ApplyEvent, "OrderCreated", map[string]any{
    "total_amt": 9900,
})

// 3. 创建事件存储和仓库
store := eventsourcing.NewGORMEventStore(gormDB)

// 可选：启用快照存储（每 50 个事件保存一次快照）
snapshotStore := eventsourcing.NewGORMSnapshotStore(gormDB)
repo, err := eventsourcing.NewRepository(store,
    func() *Order { return NewOrder("") },
    eventsourcing.WithSnapshotStore[*Order](snapshotStore),
    eventsourcing.WithSnapshotEvery[*Order](50),
)

// 不带快照的简单仓库
repo, err := eventsourcing.NewRepository(store,
    func() *Order { return NewOrder("") },
)

// 4. 保存聚合（持久化未提交事件）
if err := repo.Save(ctx, order); err != nil { ... }

// 5. 加载聚合（重放所有历史事件）
loaded, err := repo.Load(ctx, "ORD-001")
if err != nil { ... }
fmt.Println(loaded.Status, loaded.Version()) // "created", 1
```

**关键类型：**
- `eventsourcing.Aggregate` — 聚合根接口（`AggregateID`, `AggregateType`, `Version`, `ApplyEvent`, `UncommittedEvents`, `ClearUncommittedEvents`）
- `eventsourcing.BaseAggregate` — 可嵌入基础实现，提供 `RaiseEvent(applier, eventType, data)` 方法
- `eventsourcing.Event` — 持久化事件（`ID`, `AggregateID`, `AggregateType`, `Version`, `EventType`, `Data json.RawMessage`, `CreatedAt`）
- `eventsourcing.Snapshot` — 聚合快照（`AggregateID`, `Version`, `Data json.RawMessage`）
- `eventsourcing.Repository[T]` — 泛型仓库（`Save(ctx, aggregate)`, `Load(ctx, aggregateID)`）
- `eventsourcing.NewRepository(store, factory, opts...)` — 创建仓库
- 仓库选项：`WithSnapshotStore[T](store)`, `WithSnapshotEvery[T](n)`
- `eventsourcing.NewGORMEventStore(db)` — GORM 事件存储实现
- 错误：`ErrAggregateNotFound`, `ErrNoEvents`, `ErrNilEventStore`, `ErrNilFactory`

**注意：**
- `RaiseEvent` 需要将 `aggregate.ApplyEvent` 作为第一个参数传入（因 Go 接口限制）
- 事件存储使用 `(aggregate_id, aggregate_type, version)` 唯一约束，防止并发冲突（乐观锁）
- `Repository.Load` 先尝试快照，再加载快照后的增量事件，减少重放开销

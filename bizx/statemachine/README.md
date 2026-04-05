# bizx/statemachine — 有限状态机

适用于订单流程、工单流转、审批流等业务场景，支持**守卫条件**、**转换动作**、**状态进入/离开回调**。

## 核心类型

```go
type Transition struct {
    From   State
    Event  Event
    To     State
    Guard  func(ctx context.Context, data any) bool  // 可选：守卫条件
    Action func(ctx context.Context, data any) error // 可选：转换时执行的动作
}
```

## API

| 方法 | 说明 |
|------|------|
| `New(initial, transitions)` | 创建状态机 |
| `Fire(ctx, event, data)` | 触发事件，驱动状态转换 |
| `Current()` | 获取当前状态 |
| `Can(event)` | 当前状态是否可触发该事件 |
| `AvailableEvents()` | 当前状态可触发的事件列表 |
| `OnTransition(fn)` | 注册全局转换回调 |
| `OnEnter(state, fn)` | 注册进入某状态回调 |
| `OnLeave(state, fn)` | 注册离开某状态回调 |

## 快速上手

```go
// 定义订单状态机
sm := statemachine.New("pending", []statemachine.Transition{
    {From: "pending", Event: "pay",     To: "paid"},
    {From: "paid",    Event: "ship",    To: "shipped"},
    {From: "shipped", Event: "deliver", To: "delivered"},
    {From: "pending", Event: "cancel",  To: "cancelled"},
    {
        From:  "paid",
        Event: "refund",
        To:    "refunded",
        Guard: func(ctx context.Context, data any) bool {
            order := data.(*Order)
            return order.CanRefund() // 守卫：满足退款条件才允许
        },
        Action: func(ctx context.Context, data any) error {
            order := data.(*Order)
            return paymentService.Refund(ctx, order.ID)
        },
    },
})

// 注册回调
sm.OnEnter("paid", func(ctx context.Context, data any) {
    sendPaymentConfirmEmail(ctx, data.(*Order))
})
sm.OnTransition(func(from, to statemachine.State, event statemachine.Event) {
    log.Printf("Order: %s --[%s]--> %s", from, event, to)
})

// 驱动转换
if err := sm.Fire(ctx, "pay", order); err != nil {
    // statemachine.ErrInvalidTransition 或 statemachine.ErrGuardRejected
}

sm.Current() // "paid"
sm.Can("ship")   // true
sm.Can("cancel") // false（paid 状态没有 cancel 转换）
```

## 错误

| 错误 | 说明 |
|------|------|
| `ErrInvalidTransition` | 当前状态无此事件对应的转换 |
| `ErrGuardRejected` | 守卫条件返回 false |
| `ErrActionFailed` | Action 执行失败 |

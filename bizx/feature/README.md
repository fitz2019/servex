# bizx/feature — 特性开关

支持**百分比灰度**、**用户白名单**、**分组白名单**三种放量策略，可动态控制功能的可见性。

## 实现

| 存储构造函数 | 说明 |
|-------------|------|
| `NewMemoryStore()` | 内存存储，适合测试 |
| `NewRedisStore(client, opts...)` | Redis 存储（key 默认前缀 `feature:`） |

## 数据结构

```go
type Flag struct {
    Name       string         // 开关名称
    Enabled    bool           // 全局总开关
    Percentage int            // 百分比放量（0-100），按用户 ID hash 分桶
    Users      []string       // 用户白名单
    Groups     []string       // 分组白名单（如租户、角色）
    Metadata   map[string]any // 附加信息
}
```

## 接口

```go
type Manager interface {
    IsEnabled(ctx, name string, opts ...EvalOption) bool
    GetFlag(ctx, name string) (*Flag, error)
    SetFlag(ctx, flag *Flag) error
    DeleteFlag(ctx, name string) error
    ListFlags(ctx) ([]*Flag, error)
}
```

## 评估选项

| 选项 | 说明 |
|------|------|
| `WithUser(userID)` | 指定当前用户 |
| `WithGroup(group)` | 指定当前分组 |
| `WithAttributes(attrs)` | 附加属性（供自定义扩展） |

## 放量优先级

1. 全局开关 `Enabled = false` → 直接返回 false
2. 用户白名单命中 → 返回 true
3. 分组白名单命中 → 返回 true
4. 百分比命中（FNV hash + 100 取模）→ 返回 true/false
5. 无任何限制且全局启用 → 返回 true

## 快速上手

```go
store := feature.NewRedisStore(redisClient)
mgr := feature.NewManager(store)

// 创建特性开关（10% 灰度 + VIP 白名单）
mgr.SetFlag(ctx, &feature.Flag{
    Name:       "new_checkout_ui",
    Enabled:    true,
    Percentage: 10,
    Users:      []string{"vip_user_1", "vip_user_2"},
})

// 判断是否对当前用户启用
if mgr.IsEnabled(ctx, "new_checkout_ui", feature.WithUser(userID)) {
    renderNewUI(w)
} else {
    renderOldUI(w)
}

// 按分组放量（如对 beta 租户全量开放）
mgr.SetFlag(ctx, &feature.Flag{
    Name:    "ai_search",
    Enabled: true,
    Groups:  []string{"beta_tenants"},
})

enabled := mgr.IsEnabled(ctx, "ai_search",
    feature.WithUser(userID),
    feature.WithGroup(tenantGroup),
)
```

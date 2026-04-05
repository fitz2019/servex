# bizx/sequence — 业务序号生成器

生成连续有意义的业务编号（如订单号、单据号），区别于 `xutil/idgen` 的全局唯一 ID，本包支持**前缀**、**日期**、**补零**、**每日重置**等业务格式。

## 实现

| 存储构造函数 | 说明 |
|-------------|------|
| `NewMemoryStore()` | 内存存储，适合测试或单进程场景 |
| `NewRedisStore(client)` | Redis 存储，支持分布式高并发 |

## 接口

```go
type Sequence interface {
    Next(ctx) (string, error)    // 生成下一个序号
    Current(ctx) (string, error) // 获取当前序号（不递增）
    Reset(ctx) error             // 重置
}
```

## 配置

```go
type Config struct {
    Name       string // 序列名（如 "order"）
    Prefix     string // 前缀（如 "ORD-"）
    DateFormat string // 日期格式（如 "20060102"），为空则不含日期
    Padding    int    // 序号补零位数，默认 4 → 0001
    Step       int64  // 步长，默认 1
    ResetDaily bool   // 每日重置，默认 false
}
```

## 快速上手

```go
store := sequence.NewRedisStore(redisClient)
seq := sequence.New(&sequence.Config{
    Name:       "order",
    Prefix:     "ORD-",
    DateFormat: "20060102",
    Padding:    4,
    ResetDaily: true, // 每天从 0001 重新开始
}, store)

id, _ := seq.Next(ctx) // "ORD-20260405-0001"
id, _ = seq.Next(ctx)  // "ORD-20260405-0002"
```

## 不含日期的序号

```go
seq := sequence.New(&sequence.Config{
    Name:    "invoice",
    Prefix:  "INV-",
    Padding: 6,
}, store)

id, _ := seq.Next(ctx) // "INV-000001"
id, _ = seq.Next(ctx)  // "INV-000002"
```

## 注意事项

- Redis 存储使用 `INCRBY` 保证原子性，天然支持高并发
- `ResetDaily: true` 时，不同日期的 key 独立计数（按日期后缀区分），旧 key 不会自动清理

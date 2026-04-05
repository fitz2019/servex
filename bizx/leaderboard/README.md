# bizx/leaderboard — 排行榜

提供语义化排行榜接口，区别于 `storage/redis` 的 ZSet 原始操作，本包封装了**分页**、**并列排名**、**升降序**等业务功能。

## 实现

| 构造函数 | 说明 |
|----------|------|
| `NewMemoryLeaderboard(name, opts...)` | 内存实现，适合测试 |
| `NewRedisLeaderboard(client, name, opts...)` | Redis Sorted Set 实现 |

## 接口

```go
type Leaderboard interface {
    AddScore(ctx, member string, score float64) error          // 设置分数（覆盖）
    IncrScore(ctx, member string, delta float64) (float64, error) // 增加分数
    GetRank(ctx, member string) (*Entry, error)                // 获取排名信息
    GetScore(ctx, member string) (float64, error)              // 获取分数
    TopN(ctx context.Context, n int) ([]Entry, error)          // 前 N 名
    GetPage(ctx, offset, limit int) (*Page, error)             // 分页
    Remove(ctx, members ...string) error                       // 移除成员
    Count(ctx) (int64, error)                                  // 总成员数
    Reset(ctx) error                                           // 清空
}
```

## 数据结构

```go
type Entry struct {
    Member string  // 成员标识
    Score  float64 // 分数
    Rank   int64   // 排名（1-based）
}

type Page struct {
    Entries []Entry
    Total   int64
    HasMore bool
}
```

## 选项

| 选项 | 说明 |
|------|------|
| `WithPrefix(prefix)` | key 前缀 |
| `WithOrder(Descending\|Ascending)` | 排序方式，默认降序（分数高排前） |

## 快速上手

```go
lb := leaderboard.NewRedisLeaderboard(redisClient, "weekly_score")

// 添加/更新分数
lb.AddScore(ctx, "player1", 1500)
lb.IncrScore(ctx, "player1", 200) // 累加

// 获取前 10 名
top10, _ := lb.TopN(ctx, 10)
for _, e := range top10 {
    fmt.Printf("Rank %d: %s %.0f\n", e.Rank, e.Member, e.Score)
}

// 分页获取（第 2 页，每页 20 条）
page, _ := lb.GetPage(ctx, 20, 20)

// 查询某成员排名
entry, _ := lb.GetRank(ctx, "player1")
```

# llm/serving/billing

`github.com/Tsukikage7/servex/llm/serving/billing` — AI 服务用量计费，支持按模型定价、用量记录、汇总统计及计费中间件。

## 核心类型

- `Billing` — 计费引擎接口，方法包括 Record、GetSummary、SetPricing、CalculateCost
- `PriceModel` — 定价模型，包含 ModelID、InputPricePerM、OutputPricePerM、CachedPricePerM（每百万 token 价格）
- `UsageRecord` — 单次请求用量记录，包含 KeyID、ModelID、Usage、Cost、CreatedAt
- `Summary` — 汇总统计，包含 TotalRequests、TotalTokens、TotalCost、ByModel
- `Store` — 存储接口，方法包括 SaveRecord、GetRecords、AutoMigrate
- `NewBilling(store, opts...)` — 创建计费引擎
- `NewGORMStore(db)` — 基于 GORM 的持久化存储
- `NewMemoryStore()` — 基于内存的存储（用于测试）
- `Middleware(b, keyExtractor)` — 计费中间件，在 Generate/Stream 后自动记录用量

## 使用示例

```go
import "github.com/Tsukikage7/servex/llm/serving/billing"

store := billing.NewGORMStore(db)
b := billing.NewBilling(store,
    billing.WithDefaultPricing([]billing.PriceModel{
        {ModelID: "gpt-4o", InputPricePerM: 2.5, OutputPricePerM: 10.0},
    }),
)

// 作为中间件
billingMw := billing.Middleware(b, func(ctx context.Context) string {
    key, _ := apikey.FromContext(ctx)
    if key != nil { return key.ID }
    return ""
})
billedModel := billingMw(myModel)

// 查询汇总
summary, _ := b.GetSummary(ctx, "key-123",
    time.Now().AddDate(0, -1, 0), time.Now())
fmt.Printf("总费用: $%.4f，总 tokens: %d\n",
    summary.TotalCost, summary.TotalTokens)
```

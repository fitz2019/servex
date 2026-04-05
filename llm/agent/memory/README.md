# llm/agent/memory

`github.com/Tsukikage7/servex/llm/agent/memory` — AI 对话持久化记忆，支持内存/Redis 存储后端及摘要、实体高级记忆策略。

## 核心类型

- `Store` — 记忆存储接口，方法包括 Save、Load、Delete、List
- `MemoryStore` — 基于内存的 Store 实现，线程安全
- `RedisStore` — 基于 Redis Hash 的 Store 实现，支持 TTL 和 Key 前缀
- `PersistentMemory` — 将任意 `conversation.Memory` 包装为支持持久化的记忆，额外提供 Save/Load 方法
- `SummaryMemory` — 摘要记忆，消息数超过阈值时自动调用 LLM 压缩旧消息
- `EntityMemory` — 实体记忆，自动从消息中提取命名实体并注入上下文

## 使用示例

```go
import "github.com/Tsukikage7/servex/llm/agent/memory"

// 内存 Store
store := memory.NewMemoryStore()

// Redis Store
rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
redisStore := memory.NewRedisStore(rdb,
    memory.WithKeyPrefix("myapp:memory:"),
    memory.WithTTL(24*time.Hour),
)

// 持久化记忆包装
inner := conversation.NewBufferMemory(20)
pm := memory.NewPersistentMemory(inner, store, "session-001")
_ = pm.Load(ctx)    // 从存储恢复
pm.Add(llm.UserMessage("你好"))
_ = pm.Save(ctx)    // 持久化到存储

// 摘要记忆
sm := memory.NewSummaryMemory(myModel, memory.WithMaxMessages(20))

// 实体记忆
em := memory.NewEntityMemory(myModel)
```

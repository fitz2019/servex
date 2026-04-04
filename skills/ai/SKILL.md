---
name: ai
description: servex AI 模块专家。当用户使用 servex 的 ai/openai、ai/anthropic、ai/gemini、ai/conversation、ai/embedding、ai/prompt、ai/middleware、ai/vectorstore、ai/toolcall、ai/router 时触发。
---

# servex AI

## ai/toolcall — 工具调用框架

```go
// 创建工具注册表
registry := toolcall.NewRegistry()

// 注册工具（HandlerFunc 签名：func(ctx, input) (output, error)）
registry.Register("get_weather", toolcall.HandlerFunc(
    func(ctx context.Context, input map[string]any) (any, error) {
        city := input["city"].(string)
        return map[string]any{"temp": 25, "city": city}, nil
    },
))

// 创建执行器（绑定 AI provider 和工具注册表）
executor := toolcall.NewExecutor(provider, registry,
    toolcall.WithMaxRounds(5), // 最多 5 轮工具调用
    toolcall.WithOnStep(func(event toolcall.StepEvent) {
        // 每轮工具调用的回调（可用于流式输出进度）
        fmt.Printf("工具调用: %s\n", event.ToolName)
    }),
)

// 执行（自动循环直到模型不再调用工具或达到上限）
result, err := executor.Run(ctx, messages)
if err != nil { ... }
fmt.Println("最终回答:", result.Content)
```

**关键类型：**
- `toolcall.NewRegistry()` — 工具注册表（不是 `NewTool`）
- `toolcall.NewExecutor(provider, registry, opts...)` — 执行器
- `toolcall.WithOnStep(fn)` — 步骤回调，每次工具调用后触发
- `toolcall.StepEvent` — 步骤事件，含 `ToolName`、`Input`、`Output` 字段

## ai/router — 多 Provider 路由

```go
// 创建路由器：按条件选择不同 AI Provider
router := airouter.New(
    fallbackProvider, // 默认 provider（无匹配规则时使用）
    airouter.Route{
        // 匹配条件：模型名包含 "gpt"
        Match: func(req ai.Request) bool {
            return strings.Contains(req.Model, "gpt")
        },
        Provider: openaiProvider,
    },
    airouter.Route{
        Match: func(req ai.Request) bool {
            return strings.Contains(req.Model, "claude")
        },
        Provider: anthropicProvider,
    },
)

// router 实现 ai.Provider 接口，可直接使用
resp, err := router.Generate(ctx, req)
stream, err := router.Stream(ctx, req)
```

**关键类型：**
- `airouter.New(fallback, ...Route)` — 构造器（不是 `NewRouter`）
- `airouter.Route` — 路由规则结构体，含 `Match func(ai.Request) bool` 和 `Provider ai.Provider`
- `Router` 实现 `ai.Provider` 接口，与 toolcall executor 完全兼容

## ai/openai — OpenAI 客户端

```go
// 创建 OpenAI 客户端（兼容 DeepSeek、通义千问、Azure OpenAI 等 OpenAI 格式 Provider）
client := openai.New("sk-xxx",
    openai.WithModel("gpt-4o"),                  // 默认聊天模型
    openai.WithEmbeddingModel("text-embedding-3-small"), // 默认嵌入模型
    openai.WithBaseURL("https://api.deepseek.com/v1"),   // 第三方 Provider
    openai.WithOrganization("org-xxx"),           // OpenAI 组织 ID
    openai.WithHTTPClient(customHTTPClient),      // 自定义 HTTP 客户端
)

// 非流式生成（实现 ai.ChatModel 接口）
resp, err := client.Generate(ctx, messages, ai.WithModel("gpt-4o-mini"))

// 流式生成
reader, err := client.Stream(ctx, messages)
defer reader.Close()
for {
    chunk, err := reader.Recv()
    if err == io.EOF { break }
    fmt.Print(chunk.Delta)
}

// 文本嵌入（实现 ai.EmbeddingModel 接口）
embedResp, err := client.EmbedTexts(ctx, []string{"hello", "world"})
// embedResp.Embeddings: [][]float32
```

**关键类型：**
- `openai.New(apiKey, opts...)` — 构造器，同时实现 `ai.ChatModel` 和 `ai.EmbeddingModel`
- `WithBaseURL(url)` — 兼容第三方 OpenAI 格式 Provider（DeepSeek、通义千问等）
- `WithModel(model)` / `WithEmbeddingModel(model)` — 默认模型名
- 错误自动映射：429→`ai.ErrRateLimited`，401→`ai.ErrInvalidAuth`，5xx→`ai.ErrProviderUnavailable`

## ai/anthropic — Anthropic Claude 客户端

```go
// 创建 Anthropic 客户端
client := anthropic.New("sk-ant-xxx",
    anthropic.WithModel("claude-3-5-sonnet-20241022"),
    anthropic.WithDefaultMaxTokens(4096),         // Anthropic API 必须指定 max_tokens
    anthropic.WithAnthropicVersion("2023-06-01"), // API 版本
)

// 实现 ai.ChatModel 接口
resp, err := client.Generate(ctx, messages)
reader, err := client.Stream(ctx, messages)
```

**关键类型：**
- `anthropic.New(apiKey, opts...)` — 构造器，实现 `ai.ChatModel`（不支持 Embedding）
- `WithDefaultMaxTokens(n)` — Anthropic 要求必须设置 max_tokens
- 系统消息自动提取为 Anthropic 的 `system` 字段（与 OpenAI 处理方式不同）

## ai/gemini — Google Gemini 客户端

```go
// 创建 Gemini 客户端
client := gemini.New("AIzaSy...",
    gemini.WithModel("gemini-2.0-flash"),
    gemini.WithEmbeddingModel("text-embedding-004"),
)

// 实现 ai.ChatModel + ai.EmbeddingModel 接口
resp, err := client.Generate(ctx, messages)
reader, err := client.Stream(ctx, messages)
embedResp, err := client.EmbedTexts(ctx, texts)
```

**关键类型：**
- `gemini.New(apiKey, opts...)` — 构造器，同时实现 `ai.ChatModel` 和 `ai.EmbeddingModel`
- 系统消息自动映射为 Gemini 的 `systemInstruction` 字段
- 工具调用中 ToolCallID 作为 Gemini 的函数名使用

## ai/conversation — 多轮对话会话管理

```go
// 创建会话管理器（自动维护对话历史）
conv := conversation.New(client,
    conversation.WithSystemPrompt("你是一个有帮助的助手"),
    conversation.WithMemory(conversation.NewWindowMemory(10)), // 保留最近 10 轮
)

// 发送消息（自动拼接历史 + 系统提示）
resp, err := conv.Chat(ctx, "你好")
fmt.Println(resp.Message.Content)

// 流式对话（流结束后自动记录助手消息到记忆）
reader, err := conv.ChatStream(ctx, "请继续")

// 获取历史 / 重置
history := conv.History()
conv.Reset()
```

**记忆策略：**
- `conversation.NewBufferMemory()` — 完整缓冲，保留所有历史（默认）
- `conversation.NewWindowMemory(maxRounds)` — 滑动窗口，只保留最近 N 轮
- 实现 `conversation.Memory` 接口可自定义记忆策略

## ai/embedding — 嵌入向量工具

```go
// 批量嵌入（自动分批，适合超过 Provider 单次限制的场景）
resp, err := embedding.BatchEmbed(ctx, embeddingModel, texts, 100) // batchSize=100

// 余弦相似度计算
score := embedding.CosineSimilarity(vecA, vecB) // [-1, 1]
```

**关键函数：**
- `embedding.BatchEmbed(ctx, model, texts, batchSize, opts...)` — 自动分批嵌入
- `embedding.CosineSimilarity(a, b []float32) float32` — 余弦相似度

## ai/prompt — 消息模板引擎

```go
// 创建模板（基于 Go text/template 语法）
tmpl := prompt.MustNew(ai.RoleSystem,
    "你是{{.Role}}，专注于{{.Domain}}领域。请用{{.Language}}回答。",
)

// 渲染为 ai.Message
msg := tmpl.MustRender(map[string]string{
    "Role":     "技术顾问",
    "Domain":   "Go 微服务",
    "Language": "中文",
})
// msg.Role == ai.RoleSystem, msg.Content == "你是技术顾问，..."
```

**关键类型：**
- `prompt.New(role, text) (*Template, error)` / `prompt.MustNew(role, text)` — 创建模板
- `tmpl.Render(data) (ai.Message, error)` / `tmpl.MustRender(data)` — 渲染

## ai/middleware — AI 模型中间件链

```go
// 定义中间件类型：func(ai.ChatModel) ai.ChatModel
// 内置中间件：Logging、Retry、RateLimit、UsageTracker

// 组合中间件链
wrapped := middleware.Chain(
    middleware.Logging(log),                       // 记录请求日志
    middleware.Retry(3, 500*time.Millisecond),     // 429/5xx 指数退避重试
    middleware.RateLimit(limiter),                  // 限流
    tracker.Middleware(),                           // 用量追踪
)(client) // client 是 ai.ChatModel

// 使用包装后的 model（透明，接口不变）
resp, err := wrapped.Generate(ctx, messages)

// 用量追踪
tracker := &middleware.UsageTracker{}
// ... 使用 tracker.Middleware() 包装后 ...
total := tracker.Total() // 累计 token 用量
tracker.Reset()          // 重置
```

**关键类型：**
- `middleware.Middleware` — `func(ai.ChatModel) ai.ChatModel`
- `middleware.Chain(outer, ...others)` — 链接多个中间件
- `middleware.Logging(log)` — 记录模型名、token 数、耗时
- `middleware.Retry(maxAttempts, baseDelay)` — 对 `ai.IsRetryable` 错误重试
- `middleware.RateLimit(limiter)` — 基于 `ratelimit.Limiter` 限流
- `middleware.UsageTracker` — 线程安全的 token 用量累计

## ai/vectorstore — 向量存储接口

```go
// VectorStore 统一接口
type VectorStore interface {
    AddDocuments(ctx context.Context, docs []Document) error
    SimilaritySearch(ctx context.Context, query []float32, k int, opts ...SearchOption) ([]SearchResult, error)
    Delete(ctx context.Context, ids []string) error
}

// 使用示例
store.AddDocuments(ctx, []vectorstore.Document{
    {ID: "1", Content: "Go 微服务", Vector: vec1, Metadata: map[string]any{"topic": "tech"}},
})

results, err := store.SimilaritySearch(ctx, queryVec, 5,
    vectorstore.WithFilter(map[string]any{"topic": "tech"}),
    vectorstore.WithScoreThreshold(0.8),
)
for _, r := range results {
    fmt.Printf("ID=%s Score=%.2f Content=%s\n", r.Document.ID, r.Score, r.Document.Content)
}
```

**关键类型：**
- `vectorstore.VectorStore` — 向量存储统一接口
- `vectorstore.Document` — 文档（ID、Content、Vector、Metadata）
- `vectorstore.SearchResult` — 搜索结果（Document + Score）
- `vectorstore.WithFilter(map)` — 元数据过滤
- `vectorstore.WithScoreThreshold(float32)` — 相似度阈值过滤

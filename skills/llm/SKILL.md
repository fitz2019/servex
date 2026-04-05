---
name: llm
description: servex LLM 模块专家。当用户使用 servex 的 llm/provider/openai、llm/provider/anthropic、llm/provider/gemini、llm/agent/conversation、llm/retrieval/embedding、llm/prompt、llm/middleware、llm/retrieval/vectorstore、llm/agent/toolcall、llm/provider/router、llm/retrieval/splitter、llm/processing/structured、llm/serving/cache、llm/safety/guardrail、llm/retrieval/rag、llm/agent/chain、llm/retrieval/document、llm/agent/memory、llm/retrieval/rerank、llm/agent、llm/eval、llm/processing/tokenizer、llm/safety/moderation、llm/serving/apikey、llm/serving/billing、llm/serving/proxy、llm/processing/classifier、llm/processing/extractor、llm/processing/translator 时触发。
---

# servex LLM

## llm/agent/toolcall — 工具调用框架

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

## llm/provider/router — 多 Provider 路由

```go
// 创建路由器：按条件选择不同 AI Provider
router := airouter.New(
    fallbackProvider, // 默认 provider（无匹配规则时使用）
    airouter.Route{
        // 匹配条件：模型名包含 "gpt"
        Match: func(req llm.Request) bool {
            return strings.Contains(req.Model, "gpt")
        },
        Provider: openaiProvider,
    },
    airouter.Route{
        Match: func(req llm.Request) bool {
            return strings.Contains(req.Model, "claude")
        },
        Provider: anthropicProvider,
    },
)

// router 实现 llm.Provider 接口，可直接使用
resp, err := router.Generate(ctx, req)
stream, err := router.Stream(ctx, req)
```

**关键类型：**
- `airouter.New(fallback, ...Route)` — 构造器（不是 `NewRouter`）
- `airouter.Route` — 路由规则结构体，含 `Match func(llm.Request) bool` 和 `Provider llm.Provider`
- `Router` 实现 `llm.Provider` 接口，与 toolcall executor 完全兼容

## llm/provider/openai — OpenAI 客户端

```go
// 创建 OpenAI 客户端（兼容 DeepSeek、通义千问、Azure OpenAI 等 OpenAI 格式 Provider）
client := openai.New("sk-xxx",
    openllm.WithModel("gpt-4o"),                  // 默认聊天模型
    openai.WithEmbeddingModel("text-embedding-3-small"), // 默认嵌入模型
    openai.WithBaseURL("https://api.deepseek.com/v1"),   // 第三方 Provider
    openai.WithOrganization("org-xxx"),           // OpenAI 组织 ID
    openai.WithHTTPClient(customHTTPClient),      // 自定义 HTTP 客户端
)

// 非流式生成（实现 llm.ChatModel 接口）
resp, err := client.Generate(ctx, messages, llm.WithModel("gpt-4o-mini"))

// 流式生成
reader, err := client.Stream(ctx, messages)
defer reader.Close()
for {
    chunk, err := reader.Recv()
    if err == io.EOF { break }
    fmt.Print(chunk.Delta)
}

// 文本嵌入（实现 llm.EmbeddingModel 接口）
embedResp, err := client.EmbedTexts(ctx, []string{"hello", "world"})
// embedResp.Embeddings: [][]float32
```

**关键类型：**
- `openai.New(apiKey, opts...)` — 构造器，同时实现 `llm.ChatModel` 和 `llm.EmbeddingModel`
- `WithBaseURL(url)` — 兼容第三方 OpenAI 格式 Provider（DeepSeek、通义千问等）
- `WithModel(model)` / `WithEmbeddingModel(model)` — 默认模型名
- 错误自动映射：429→`ai.ErrRateLimited`，401→`ai.ErrInvalidAuth`，5xx→`ai.ErrProviderUnavailable`

## llm/provider/anthropic — Anthropic Claude 客户端

```go
// 创建 Anthropic 客户端
client := anthropic.New("sk-ant-xxx",
    anthropic.WithModel("claude-3-5-sonnet-20241022"),
    anthropic.WithDefaultMaxTokens(4096),         // Anthropic API 必须指定 max_tokens
    anthropic.WithAnthropicVersion("2023-06-01"), // API 版本
)

// 实现 llm.ChatModel 接口
resp, err := client.Generate(ctx, messages)
reader, err := client.Stream(ctx, messages)
```

**关键类型：**
- `anthropic.New(apiKey, opts...)` — 构造器，实现 `llm.ChatModel`（不支持 Embedding）
- `WithDefaultMaxTokens(n)` — Anthropic 要求必须设置 max_tokens
- 系统消息自动提取为 Anthropic 的 `system` 字段（与 OpenAI 处理方式不同）

## llm/provider/gemini — Google Gemini 客户端

```go
// 创建 Gemini 客户端
client := gemini.New("AIzaSy...",
    gemini.WithModel("gemini-2.0-flash"),
    gemini.WithEmbeddingModel("text-embedding-004"),
)

// 实现 llm.ChatModel + llm.EmbeddingModel 接口
resp, err := client.Generate(ctx, messages)
reader, err := client.Stream(ctx, messages)
embedResp, err := client.EmbedTexts(ctx, texts)
```

**关键类型：**
- `gemini.New(apiKey, opts...)` — 构造器，同时实现 `llm.ChatModel` 和 `llm.EmbeddingModel`
- 系统消息自动映射为 Gemini 的 `systemInstruction` 字段
- 工具调用中 ToolCallID 作为 Gemini 的函数名使用

## llm/agent/conversation — 多轮对话会话管理

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

## llm/retrieval/embedding — 嵌入向量工具

```go
// 批量嵌入（自动分批，适合超过 Provider 单次限制的场景）
resp, err := embedding.BatchEmbed(ctx, embeddingModel, texts, 100) // batchSize=100

// 余弦相似度计算
score := embedding.CosineSimilarity(vecA, vecB) // [-1, 1]
```

**关键函数：**
- `embedding.BatchEmbed(ctx, model, texts, batchSize, opts...)` — 自动分批嵌入
- `embedding.CosineSimilarity(a, b []float32) float32` — 余弦相似度

## llm/prompt — 消息模板引擎

```go
// 创建模板（基于 Go text/template 语法）
tmpl := prompt.MustNew(llm.RoleSystem,
    "你是{{.Role}}，专注于{{.Domain}}领域。请用{{.Language}}回答。",
)

// 渲染为 llm.Message
msg := tmpl.MustRender(map[string]string{
    "Role":     "技术顾问",
    "Domain":   "Go 微服务",
    "Language": "中文",
})
// msg.Role == llm.RoleSystem, msg.Content == "你是技术顾问，..."
```

**关键类型：**
- `prompt.New(role, text) (*Template, error)` / `prompt.MustNew(role, text)` — 创建模板
- `tmpl.Render(data) (llm.Message, error)` / `tmpl.MustRender(data)` — 渲染

## llm/middleware — AI 模型中间件链

```go
// 定义中间件类型：func(llm.ChatModel) llm.ChatModel
// 内置中间件：Logging、Retry、RateLimit、UsageTracker

// 组合中间件链
wrapped := middleware.Chain(
    middleware.Logging(log),                       // 记录请求日志
    middleware.Retry(3, 500*time.Millisecond),     // 429/5xx 指数退避重试
    middleware.RateLimit(limiter),                  // 限流
    tracker.Middleware(),                           // 用量追踪
)(client) // client 是 llm.ChatModel

// 使用包装后的 model（透明，接口不变）
resp, err := wrapped.Generate(ctx, messages)

// 用量追踪
tracker := &middleware.UsageTracker{}
// ... 使用 tracker.Middleware() 包装后 ...
total := tracker.Total() // 累计 token 用量
tracker.Reset()          // 重置
```

**关键类型：**
- `middleware.Middleware` — `func(llm.ChatModel) llm.ChatModel`
- `middleware.Chain(outer, ...others)` — 链接多个中间件
- `middleware.Logging(log)` — 记录模型名、token 数、耗时
- `middleware.Retry(maxAttempts, baseDelay)` — 对 `ai.IsRetryable` 错误重试
- `middleware.RateLimit(limiter)` — 基于 `ratelimit.Limiter` 限流
- `middleware.UsageTracker` — 线程安全的 token 用量累计

## llm/retrieval/vectorstore — 向量存储接口

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

## llm/retrieval/splitter — 文本分块器

**接口与类型：**

```go
// Splitter 分块器接口
type Splitter interface {
    Split(text string) []Chunk
}

// Chunk 文本块
type Chunk struct {
    Text     string         // 块文本
    Offset   int            // 原文起始字节偏移
    Index    int            // 块序号（从 0 开始）
    Metadata map[string]any // 附加元数据
}
```

**三种实现：**
- `NewCharacterSplitter(opts...)` — 按字符数硬切，保留相邻块 chunkOverlap 字符重叠
- `NewRecursiveSplitter(opts...)` — 递归按分隔符（段落→句子→字符）拆分后合并，默认分隔符：`["\n\n", "\n", "。", ".", " ", ""]`
- `NewTokenSplitter(opts...)` — 按估算 Token 数分块（ASCII 4字符≈1 token，CJK 1.5字符≈1 token）

**通用选项：**
- `WithChunkSize(n)` — 每块最大字符/Token 数（默认 1000）
- `WithChunkOverlap(n)` — 相邻块重叠字符/Token 数（默认 200）
- `WithSeparators([]string)` — 覆盖递归分隔符列表（仅 RecursiveSplitter）

```go
// 按递归分隔符分块，每块最多 500 字符，重叠 50 字符
sp := splitter.NewRecursiveSplitter(
    splitter.WithChunkSize(500),
    splitter.WithChunkOverlap(50),
)
chunks := sp.Split(longText)
for _, c := range chunks {
    fmt.Printf("[%d] offset=%d len=%d\n", c.Index, c.Offset, len([]rune(c.Text)))
}
```

## llm/processing/structured — 结构化输出提取

**核心函数：**
- `Extract[T any](ctx, model, prompt, opts...) (T, error)` — 从单条 prompt 提取结构化数据
- `ExtractFromMessages[T any](ctx, model, messages, opts...) (T, error)` — 从已有消息列表提取
- `SchemaFrom[T any]() json.RawMessage` — 从 Go struct 泛型参数生成 JSON Schema

**选项：**
- `WithMaxRetries(n)` — JSON 解析失败时最大重试次数（默认 3）
- `WithCallOptions(opts...)` — 透传底层模型调用选项
- `WithSchemaDescription(desc)` — 设置 Schema 顶层 description

```go
type Article struct {
    Title    string   `json:"title"    description:"文章标题"`
    Tags     []string `json:"tags"     description:"关键词列表"`
    Sentiment string  `json:"sentiment" description:"情感倾向：positive/negative/neutral"`
}

article, err := structured.Extract[Article](ctx, model,
    "请分析以下新闻："+newsText,
    structured.WithMaxRetries(2),
)
if err != nil { ... }
fmt.Println(article.Title, article.Sentiment)
```

**说明：** `SchemaFrom` 支持 string/int/float/bool/slice/嵌套 struct，使用 `json` tag 作为字段名、`description` tag 作为字段描述。解析失败时自动将错误追加到消息历史引导模型修正。

## llm/serving/cache — 语义缓存

**核心类型：**

```go
type Config struct {
    EmbeddingModel llm.EmbeddingModel // 用于将查询文本转向量（必填）
    Store          Store             // 缓存存储后端（必填）
    Threshold      float32           // 相似度阈值，超过则命中（默认 0.95）
    TTL            time.Duration     // 缓存存活时长（默认 1h）
}

type Store interface {
    Put(ctx context.Context, key []float32, value *llm.ChatResponse, ttl time.Duration) error
    Search(ctx context.Context, query []float32, threshold float32) (*llm.ChatResponse, error)
    Clear(ctx context.Context) error
}
```

**内置实现：**
- `NewMemoryStore()` — 内存存储，线性扫描余弦相似度，自动清理过期条目

**使用方式：**
- `Middleware(cfg)` — 返回 `llm.Middleware`，对 Generate 和 Stream 均生效，可加入中间件链
- `NewCachedModel(model, cfg)` — 直接包装模型，返回带缓存的 `llm.ChatModel`

```go
store := cache.NewMemoryStore()
cachedModel := cache.NewCachedModel(baseModel, &cache.Config{
    EmbeddingModel: embedModel,
    Store:          store,
    Threshold:      0.92,
    TTL:            30 * time.Minute,
})

// 与普通 ChatModel 用法完全一致，相似问题自动命中缓存
resp, err := cachedModel.Generate(ctx, messages)
```

## llm/safety/guardrail — 输入输出护栏

**Guard 接口：**

```go
type Guard interface {
    Check(ctx context.Context, messages []llm.Message) error
}
// GuardFunc 函数适配器，实现 Guard 接口
type GuardFunc func(ctx context.Context, messages []llm.Message) error
```

**内置 Guard：**
- `MaxLength(maxChars int)` — 所有消息总字符数超限返回 `ErrTooLong`
- `MaxMessages(n int)` — 消息数量超限返回 `ErrTooMany`
- `KeywordFilter(keywords []string)` — 大小写不敏感关键词过滤，命中返回 `ErrBlocked`
- `RegexFilter(patterns []string)` — 正则过滤，命中返回 `ErrBlocked`
- `PIIDetector(patterns ...PIIPattern)` — PII 检测，内置 `PIIEmail`/`PIIPhone`/`PIIIDCard`/`PIICreditCard`，命中返回 `ErrPIIDetected`

**错误常量：** `ErrBlocked`、`ErrPIIDetected`、`ErrTooLong`、`ErrTooMany`

**Middleware 构造：**
- `WithInputGuards(guards...)` — 在调用模型前执行
- `WithOutputGuards(guards...)` — 在模型返回后对响应执行（仅 Generate，流式不适用）

```go
mw := guardrail.Middleware(
    guardrail.WithInputGuards(
        guardrail.KeywordFilter([]string{"违禁词1", "违禁词2"}),
        guardrail.PIIDetector(guardrail.PIIPhone, guardrail.PIIIDCard),
        guardrail.MaxLength(4000),
    ),
    guardrail.WithOutputGuards(
        guardrail.MaxLength(8000),
    ),
)
safeModel := mw(baseModel)
resp, err := safeModel.Generate(ctx, messages)
if errors.Is(err, guardrail.ErrBlocked) {
    // 输入被拦截
}
```

## llm/retrieval/rag — RAG 管线

**核心类型：**

```go
type Config struct {
    ChatModel      llm.ChatModel           // 聊天模型（必填）
    EmbeddingModel llm.EmbeddingModel      // 嵌入模型（必填）
    VectorStore    vectorstore.VectorStore // 向量存储（必填）
    Splitter       splitter.Splitter      // 文本分块器（可选）
    TopK           int                    // 检索数量（默认 5）
    ScoreThreshold float32                // 最低相关度（默认 0）
    PromptTemplate *prompt.Template       // 自定义提示词模板（可选）
}

type Document struct {
    ID       string
    Content  string
    Metadata map[string]any
}

type Result struct {
    Answer  string         // 模型生成的回答
    Sources []RetrievedDoc // 用于生成回答的检索文档
    Usage   llm.Usage       // token 用量
}
```

**Pipeline 方法：**
- `New(cfg) (*Pipeline, error)` — 创建管线，验证必填字段
- `Ingest(ctx, docs)` — 导入文档：分块（可选）→ 批量嵌入 → 存入向量库
- `Retrieve(ctx, question)` — 只检索，返回 `[]RetrievedDoc`
- `Query(ctx, question, opts...)` — 检索增强生成（非流式），返回 `*Result`
- `QueryStream(ctx, question, opts...)` — 流式检索增强生成，返回 `llm.StreamReader`

```go
pipeline, err := rag.New(&rag.Config{
    ChatModel:      chatModel,
    EmbeddingModel: embedModel,
    VectorStore:    memStore,
    Splitter:       splitter.NewRecursiveSplitter(splitter.WithChunkSize(500)),
    TopK:           3,
    ScoreThreshold: 0.7,
})
if err != nil { ... }

// 导入文档
_ = pipeline.Ingest(ctx, []rag.Document{
    {ID: "doc1", Content: "Go 是 Google 开发的编译型语言..."},
})

// 检索增强生成
result, err := pipeline.Query(ctx, "Go 语言的特点是什么？")
fmt.Println(result.Answer)
for _, src := range result.Sources {
    fmt.Printf("来源: %s (%.2f)\n", src.ID, src.Score)
}
```

## llm/agent/chain — 多步 LLM 编排

**核心类型：**

```go
type Step struct {
    Name   string           // 步骤名称
    Prompt *prompt.Template // 渲染 prompt（必填）
    Model  llm.ChatModel     // 可选，覆盖链默认模型
    Parser func(string) (any, error) // 可选，解析输出
}

type Result struct {
    Output string        // 最后一步原始输出
    Parsed any           // 最后一步解析结果（无 Parser 时为 nil）
    Steps  []StepResult  // 各步骤结果
    Usage  llm.Usage      // 所有步骤累计 token 用量
}
```

**构造与执行：**
- `New(opts...)` — 创建编排链
- `WithModel(model)` — 设置链的默认模型
- `WithLogger(log)` — 设置日志记录器（记录每步耗时和 token）
- `WithOnStep(fn)` — 设置步骤完成回调 `func(ctx, StepEvent)`
- `(c *Chain) AddStep(step) *Chain` — 追加步骤，支持链式调用
- `(c *Chain) Run(ctx, input, opts...) (*Result, error)` — 顺序执行所有步骤

**执行规则：** 第一步用 `input` 渲染 Prompt；后续步骤用 `map{"Input": 原始 input, "Previous": 上一步输出}` 渲染。若步骤有 Parser，其结果作为下一步输入，否则使用原始字符串输出。

```go
summarize := prompt.MustNew(llm.RoleUser, "请将以下文本摘要为 3 句话：\n{{.}}")
translate := prompt.MustNew(llm.RoleUser,
    "将以下摘要翻译为英文：\n{{.Previous}}")

c := chain.New(chain.WithModel(model)).
    AddStep(chain.Step{Name: "summarize", Prompt: summarize}).
    AddStep(chain.Step{Name: "translate", Prompt: translate})

result, err := c.Run(ctx, longChineseText)
if err != nil { ... }
fmt.Println("英文摘要:", result.Output)
fmt.Printf("共 %d steps，总 tokens=%d\n", len(result.Steps), result.Usage.TotalTokens)
```

## llm/retrieval/document — 文档加载器

**Loader 接口：**

```go
type Loader interface {
    Load(ctx context.Context) ([]rag.Document, error)
}
```

**5 种加载器：**
- `NewTextLoader(reader, opts...)` — 从 `io.Reader` 读取全部内容，返回单个 Document
- `NewTextFileLoader(path, opts...)` — 从文件路径读取全部内容，返回单个 Document
- `NewCSVLoader(reader, opts...)` — 读取 CSV，每行生成一个 Document（内容列默认 "content"，其余列作元数据）
- `NewJSONLoader(reader, opts...)` — 读取 JSON 数组或 JSONL，每条对象生成一个 Document（内容字段默认 "content"）
- `NewMarkdownLoader(reader, opts...)` — 按 `##` / `###` 标题分节，每节生成一个 Document（标题存入元数据）
- `NewDirectoryLoader(dir, glob, opts...)` — 遍历目录，按 glob 模式匹配文件，每文件作为文本 Document 加载

**通用选项：**
- `WithMetadata(map[string]any)` — 附加用户自定义元数据（与加载器生成的元数据合并，用户优先）
- `WithIDPrefix(string)` — 设置 Document ID 前缀

**CSV 专属选项：**
- `WithCSVContentColumn(col)` — 指定内容列名（默认 "content"）
- `WithCSVMetadataColumns(cols...)` — 指定元数据列名列表（默认除内容列外全部）

**JSON 专属选项：**
- `WithJSONContentField(field)` — 指定内容字段名（默认 "content"）
- `WithJSONMetadataFields(fields...)` — 指定元数据字段名列表

**错误常量：** `ErrEmptyContent`、`ErrInvalidFormat`

```go
// 加载文本文件
loader := document.NewTextFileLoader("/path/to/doc.txt",
    document.WithMetadata(map[string]any{"source_type": "manual"}),
    document.WithIDPrefix("doc-"),
)
docs, err := loader.Load(ctx)

// 加载 CSV（自定义内容列）
f, _ := os.Open("data.csv")
csvLoader := document.NewCSVLoader(f,
    document.WithCSVContentColumn("body"),
    document.WithCSVMetadataColumns("title", "author"),
)
docs, err = csvLoader.Load(ctx)

// 加载 JSON 数组
jsonLoader := document.NewJSONLoader(strings.NewReader(`[{"content":"hello","tag":"go"}]`),
    document.WithJSONMetadataFields("tag"),
)
docs, err = jsonLoader.Load(ctx)

// 加载 Markdown（按标题分节）
mdLoader := document.NewMarkdownLoader(strings.NewReader(markdownText))
docs, err = mdLoader.Load(ctx)
// docs[0].Metadata["heading"] == "节标题"

// 遍历目录加载所有 .txt 文件
dirLoader := document.NewDirectoryLoader("/docs", "*.txt")
docs, err = dirLoader.Load(ctx)
```

## llm/agent/memory — 持久化记忆

**Store 接口：**

```go
type Store interface {
    Save(ctx context.Context, sessionID string, messages []llm.Message, metadata map[string]any) error
    Load(ctx context.Context, sessionID string) ([]llm.Message, map[string]any, error)
    Delete(ctx context.Context, sessionID string) error
    List(ctx context.Context) ([]string, error)
}
```

**内置实现：**

- `NewMemoryStore()` — 基于内存的线程安全存储，适合开发/测试
- `NewRedisStore(client, opts...)` — 基于 Redis Hash 的存储，数据序列化为 JSON，支持 TTL
  - `WithKeyPrefix(prefix)` — Redis Key 前缀（默认 `"servex:memory:"`）
  - `WithTTL(duration)` — 过期时间（默认 24h）
- `NewPersistentMemory(inner, store, sessionID)` — 将任意 `conversation.Memory` 包装为支持持久化的记忆，提供 `Save(ctx)` / `Load(ctx)` 方法
- `NewSummaryMemory(model, opts...)` — 摘要记忆：消息数超过阈值时自动调用 LLM 压缩旧消息为摘要，`Messages()` 返回 `[摘要系统消息] + 近期消息`
  - `WithMaxMessages(n)` — 触发摘要的阈值（默认 20）
  - `WithSummaryPrompt(p)` — 自定义摘要提示词
- `NewEntityMemory(model, opts...)` — 实体记忆：每次 `Add` 时异步从消息中抽取命名实体（人名/地名/组织等），`Messages()` 注入已知实体上下文
  - `WithEntityPrompt(p)` — 自定义实体抽取提示词

**错误常量：** `ErrSessionNotFound`、`ErrNilStore`、`ErrNilModel`

```go
// 基于 Redis 的持久化记忆
rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
store := memory.NewRedisStore(rdb,
    memory.WithKeyPrefix("myapp:session:"),
    memory.WithTTL(12*time.Hour),
)

// 用 PersistentMemory 包装 WindowMemory
inner := conversation.NewWindowMemory(20)
pm := memory.NewPersistentMemory(inner, store, "user-123")
_ = pm.Load(ctx) // 从 Redis 恢复历史消息
pm.Add(llm.UserMessage("你好"))
pm.Add(llm.AssistantMessage("你好！有什么可以帮您？"))
_ = pm.Save(ctx) // 持久化到 Redis

// 摘要记忆
sm := memory.NewSummaryMemory(chatModel,
    memory.WithMaxMessages(10),
)
sm.Add(llm.UserMessage("第一条消息"))
// 消息超过 10 条后自动摘要压缩

// 实体记忆
em := memory.NewEntityMemory(chatModel)
em.Add(llm.UserMessage("张三在北京的阿里巴巴工作"))
// 后续 em.Messages() 会自动注入：已知实体：张三: ...; 北京: ...; 阿里巴巴: ...
entities := em.Entities() // map[string]string 实体快照
```

## llm/retrieval/rerank — 重排序器

**Reranker 接口：**

```go
type Reranker interface {
    Rerank(ctx context.Context, query string, docs []rag.RetrievedDoc) ([]rag.RetrievedDoc, error)
}
```

**3 种实现：**

- `NewLLMReranker(model, opts...)` — 调用 LLM 对文档按相关性评分（0-10），分批处理后按分数降序排列
  - `WithBatchSize(n)` — 每批文档数（默认 10）
- `NewEmbeddingReranker(model, opts...)` — 批量嵌入 query 和所有文档，按余弦相似度降序排列
- `NewCrossEncoderReranker(endpoint, opts...)` — 调用外部 Cross-Encoder API（如 Cohere Rerank），按 `relevance_score` 降序排列
  - `WithAPIKey(key)` — API 密钥（添加到 `Authorization: Bearer` 请求头）
  - `WithModel(model)` — 模型名称（透传到请求体）

**通用选项：**
- `WithTopN(n)` — 最多返回 n 个文档（0 表示全部）

**错误常量：** `ErrNilModel`、`ErrEmptyDocs`、`ErrEmptyEndpoint`、`ErrAPIFailed`

```go
// LLM 重排序
llmReranker := rerank.NewLLMReranker(chatModel,
    rerank.WithTopN(5),
    rerank.WithBatchSize(8),
)
ranked, err := llmReranker.Rerank(ctx, "Go 语言特性", retrievedDocs)

// Embedding 重排序（速度快，无需额外 API 调用）
embedReranker := rerank.NewEmbeddingReranker(embedModel,
    rerank.WithTopN(5),
)
ranked, err = embedReranker.Rerank(ctx, "Go 语言特性", retrievedDocs)

// Cross-Encoder 重排序（精度最高，需外部服务）
ceReranker := rerank.NewCrossEncoderReranker("https://api.cohere.ai/v1/rerank",
    rerank.WithAPIKey("your-cohere-key"),
    rerank.WithModel("rerank-multilingual-v3.0"),
    rerank.WithTopN(5),
)
ranked, err = ceReranker.Rerank(ctx, "Go 语言特性", retrievedDocs)

// 在 RAG 管线中使用：先检索后重排序
docs, _ := pipeline.Retrieve(ctx, query)
ranked, _ = llmReranker.Rerank(ctx, query, docs)
```

## llm/agent — 自主 Agent 框架

**核心类型：**

```go
// Agent 配置
type Config struct {
    Name          string
    Model         llm.ChatModel        // 必填
    SystemPrompt  string
    Tools         *toolcall.Registry  // 可选
    Memory        conversation.Memory // 可选
    Guardrails    []guardrail.Guard   // 可选
    Strategy      Strategy            // 默认 ReAct
    MaxIterations int                 // 默认 10
    Logger        logger.Logger       // 可选
}

// 执行结果
type Result struct {
    Output     string
    Messages   []llm.Message
    ToolCalls  []toolcall.ToolResult
    Iterations int
    Usage      llm.Usage
}
```

**Strategy 接口与内置实现：**
- `NewReActStrategy()` — 思考→行动→观察循环，底层使用 `toolcall.Executor`
- `NewPlanExecuteStrategy()` — 先让 LLM 将任务分解为步骤列表（JSON 数组），再逐步执行

**多 Agent 编排：**
- `NewSupervisor(supervisor, workers map[string]*Agent)` — 监督者模式：每个 worker 自动注册为监督者的工具，监督者通过工具调用委派任务
- `NewPipeline(agents...)` — 管道模式：串行执行，前一个的输出作为下一个的输入

**流式事件类型：** `EventThinking`、`EventToolCall`、`EventToolResult`、`EventOutput`、`EventError`

**错误常量：** `ErrNilModel`、`ErrMaxIterations`、`ErrBlocked`

```go
// 创建工具注册表
registry := toolcall.NewRegistry()
registry.Register(searchTool, func(ctx context.Context, args string) (string, error) {
    // 工具实现
    return "搜索结果", nil
})

// 创建 ReAct Agent
agent, err := agent.New(&agent.Config{
    Name:         "research-agent",
    Model:        chatModel,
    SystemPrompt: "你是一个研究助手，善于使用工具收集信息。",
    Tools:        registry,
    MaxIterations: 5,
})

// 同步执行
result, err := agent.Run(ctx, "请帮我研究 Go 泛型的最新进展")
fmt.Println(result.Output)
fmt.Printf("迭代 %d 轮，调用了 %d 个工具\n", result.Iterations, len(result.ToolCalls))

// 流式执行
ch, err := agent.RunStream(ctx, "请帮我研究 Go 泛型的最新进展")
for evt := range ch {
    switch evt.Type {
    case agent.EventThinking:
        fmt.Println("[思考]", evt.Content)
    case agent.EventToolCall:
        fmt.Println("[调用工具]", evt.ToolCall.Function.Name)
    case agent.EventOutput:
        fmt.Println("[最终输出]", evt.Content)
    }
}

// PlanExecute 策略（适合复杂多步任务）
planAgent, _ := agent.New(&agent.Config{
    Model:    chatModel,
    Strategy: agent.NewPlanExecuteStrategy(),
})
result, err = planAgent.Run(ctx, "帮我完成一份竞品分析报告")

// 监督者模式（多 Agent 协作）
researchAgent, _ := agent.New(&agent.Config{Name: "researcher", Model: chatModel})
writerAgent, _ := agent.New(&agent.Config{Name: "writer", Model: chatModel})
supervisor, _ := agent.New(&agent.Config{
    Name:  "supervisor",
    Model: chatModel,
    SystemPrompt: "你是一个项目主管，负责协调研究员和写作员完成任务。",
})
sup := agent.NewSupervisor(supervisor, map[string]*agent.Agent{
    "researcher": researchAgent,
    "writer":     writerAgent,
})
result, err = sup.Run(ctx, "写一篇关于 Go 泛型的技术博客")

// 管道模式
pipeline := agent.NewPipeline(researchAgent, writerAgent)
result, err = pipeline.Run(ctx, "Go 泛型最新进展")
```

## llm/eval — LLM 输出评估

**核心类型：**

```go
// 评估输入
type EvalInput struct {
    Question  string   // 原始问题
    Answer    string   // 待评估的答案
    Reference string   // 参考答案（正确性评估时使用）
    Context   []string // 参考资料（忠实性评估时使用）
}

// 评分结果
type Score struct {
    Name   string  // 评估维度名称
    Value  float64 // 分值 0.0-1.0
    Reason string  // 评分理由
}

type EvalResult struct {
    Scores []Score
}

// 评估器接口
type Evaluator interface {
    Evaluate(ctx context.Context, input EvalInput) (*EvalResult, error)
}
```

**4 个内置评估器（均基于 LLM）：**
- `RelevanceEvaluator(model, opts...)` — 相关性：评估回答与问题的相关程度
- `FaithfulnessEvaluator(model, opts...)` — 忠实性：评估回答是否忠实于 `EvalInput.Context`，不含虚构内容
- `CoherenceEvaluator(model, opts...)` — 连贯性：评估回答的逻辑连贯性和可读性
- `CorrectnessEvaluator(model, opts...)` — 正确性：评估回答与 `EvalInput.Reference` 的一致程度

**组合评估器：**
- `NewCompositeEvaluator(evaluators...)` — 并发运行所有子评估器，合并所有 `Score`；任一子评估器出错则整体返回该错误

**选项：**
- `WithCallOptions(opts...)` — 透传底层模型调用选项

**错误常量：** `ErrNilModel`、`ErrEmptyAnswer`、`ErrParseResponse`

```go
// 单项评估
relevance := eval.RelevanceEvaluator(chatModel)
result, err := relevance.Evaluate(ctx, eval.EvalInput{
    Question: "Go 语言的并发模型是什么？",
    Answer:   "Go 使用 goroutine 和 channel 实现并发。",
})
fmt.Printf("相关性: %.2f — %s\n", result.Scores[0].Value, result.Scores[0].Reason)

// 忠实性评估（需提供参考资料）
faithfulness := eval.FaithfulnessEvaluator(chatModel)
result, err = faithfulness.Evaluate(ctx, eval.EvalInput{
    Question: "Go 的垃圾回收机制？",
    Answer:   "Go 使用三色标记清除算法进行垃圾回收。",
    Context: []string{
        "Go 运行时包含一个并发的垃圾回收器，基于三色标记清除算法。",
    },
})

// 正确性评估（需提供参考答案）
correctness := eval.CorrectnessEvaluator(chatModel)
result, err = correctness.Evaluate(ctx, eval.EvalInput{
    Question:  "1+1=?",
    Answer:    "2",
    Reference: "2",
})

// 组合评估器（并发执行全部维度）
composite := eval.NewCompositeEvaluator(
    eval.RelevanceEvaluator(chatModel),
    eval.FaithfulnessEvaluator(chatModel),
    eval.CoherenceEvaluator(chatModel),
)
result, err = composite.Evaluate(ctx, eval.EvalInput{
    Question: "Go 泛型的主要用途？",
    Answer:   "Go 泛型允许编写类型参数化的函数和数据结构，提高代码复用性。",
    Context:  []string{"Go 1.18 引入泛型，支持类型参数..."},
})
for _, s := range result.Scores {
    fmt.Printf("%s: %.2f\n", s.Name, s.Value)
}
```

## llm/processing/tokenizer — Token 计数与截断

**Tokenizer 接口：**

```go
type Tokenizer interface {
    Count(text string) int                        // 估算文本 Token 数
    CountMessages(messages []llm.Message) int      // 估算消息列表总 Token 数（含每条消息固定开销）
    Truncate(text string, maxTokens int) string   // 截断文本至不超过 maxTokens 个 Token
}
```

**两种实现：**
- `NewEstimateTokenizer(opts...)` — 基于字符比例的估算 Tokenizer，可自定义参数
- `NewCL100KTokenizer()` — 预置 CL100K 参数（GPT-4/Claude 系列），英文 4 chars/token，CJK 1.5 chars/token，每条消息固定开销 4 tokens

**构造选项（仅 EstimateTokenizer）：**
- `WithCharsPerToken(n float64)` — ASCII 字符每 Token 占用的字符数（默认 4.0）
- `WithCJKCharsPerToken(n float64)` — CJK 字符每 Token 占用的字符数（默认 1.5）
- `WithOverheadPerMessage(n int)` — 每条消息的固定 Token 开销（默认 4）

**包级辅助函数（使用默认 CL100K Tokenizer）：**
- `EstimateTokens(text string) int` — 估算文本 Token 数
- `EstimateMessageTokens(messages []llm.Message) int` — 估算消息列表总 Token 数
- `FitsContext(messages []llm.Message, maxTokens int) bool` — 判断消息列表是否在上下文窗口内
- `TruncateToFit(messages []llm.Message, maxTokens int) []llm.Message` — 截断消息列表，保留系统消息，从最旧的非系统消息依次丢弃

```go
// 使用包级函数（最常用）
tokens := tokenizer.EstimateTokens("Hello, 世界！")

ok := tokenizer.FitsContext(messages, 8192)

// 上下文窗口不足时截断，保留系统消息
trimmed := tokenizer.TruncateToFit(messages, 4096)

// 自定义 Tokenizer
tok := tokenizer.NewEstimateTokenizer(
    tokenizer.WithCharsPerToken(3.5),
    tokenizer.WithCJKCharsPerToken(1.2),
)
n := tok.Count("这是一段中文文本")

// CL100K Tokenizer（GPT-4 / Claude 估算）
cl100k := tokenizer.NewCL100KTokenizer()
n = cl100k.CountMessages([]llm.Message{
    llm.SystemMessage("你是一个助手"),
    llm.UserMessage("你好"),
})
truncated := cl100k.Truncate(longText, 512)
```

## llm/safety/moderation — 内容审核

**核心类型：**

```go
// 审核类别常量
const (
    CategoryViolence  Category = "violence"
    CategorySexual    Category = "sexual"
    CategoryHate      Category = "hate"
    CategorySelfHarm  Category = "self_harm"
    CategoryDangerous Category = "dangerous"
    CategoryPolitical Category = "political"
    CategorySpam      Category = "spam"
)

// 审核结果
type Result struct {
    Flagged    bool                 // 是否被标记为违规
    Categories map[Category]bool   // 各类别是否命中
    Scores     map[Category]float64 // 各类别置信度（0.0～1.0）
    Reason     string               // 审核理由
}

// Moderator 接口
type Moderator interface {
    Moderate(ctx context.Context, text string) (*Result, error)
    ModerateMessages(ctx context.Context, messages []llm.Message) (*Result, error)
}
```

**3 种实现：**
- `NewLLMModerator(model, opts...)` — 基于 ChatModel 的 LLM 审核器，将文本连同分类列表发给模型，解析 JSON 响应
- `NewKeywordModerator(rules map[Category][]string)` — 关键词匹配审核器，大小写不敏感，命中时分数为 1.0
- `NewCompositeModerator(moderators...Moderator)` — 组合审核器，顺序执行各审核器并合并结果（取各类别最高分）；若关键词审核已触发标记，跳过后续 LLM 审核器（短路）

**选项（LLMModerator）：**
- `WithThreshold(t float64)` — 触发标记的分数阈值（默认 0.7）
- `WithCategories(cats ...Category)` — 设置检测的类别子集（默认全部）

**错误常量：** `ErrNilModel`、`ErrEmptyText`

```go
// 关键词审核（轻量快速）
kwMod := moderation.NewKeywordModerator(map[moderation.Category][]string{
    moderation.CategoryViolence: {"暴力", "伤害"},
    moderation.CategorySpam:    {"广告", "点击链接"},
})
result, err := kwMod.Moderate(ctx, userInput)
if result.Flagged {
    fmt.Println("违规:", result.Reason)
}

// LLM 审核（语义级别，精度更高）
llmMod := moderation.NewLLMModerator(chatModel,
    moderation.WithThreshold(0.8),
    moderation.WithCategories(moderation.CategoryViolence, moderation.CategoryHate),
)
result, err = llmMod.Moderate(ctx, userInput)

// 组合审核（关键词先行，通过后再 LLM）
mod := moderation.NewCompositeModerator(kwMod, llmMod)
result, err = mod.ModerateMessages(ctx, messages)
if result.Flagged {
    fmt.Printf("命中类别: %v\n", result.Categories)
}
```

## llm/serving/apikey — API Key 管理

**核心类型：**

```go
// Key 模型（GORM 表：api_keys）
type Key struct {
    ID          string     // UUID
    Name        string
    HashedKey   string     // SHA-256 哈希（不对外暴露）
    Prefix      string     // 前缀 + 前 8 个十六进制字符（用于展示）
    OwnerID     string
    Permissions []string
    RateLimit   int        // 每分钟请求数上限（0 = 不限）
    QuotaLimit  int64      // Token 配额上限（0 = 不限）
    QuotaUsed   int64
    ExpiresAt   *time.Time
    Enabled     bool
    CreatedAt   time.Time
    LastUsedAt  *time.Time
}

// Manager 接口
type Manager interface {
    Create(ctx, opts ...CreateOption) (rawKey string, key *Key, err error)
    Validate(ctx, rawKey string) (*Key, error)
    Revoke(ctx, keyID string) error
    List(ctx, ownerID string) ([]*Key, error)
    UpdateQuota(ctx, keyID string, tokensUsed int64) error
}

// Store 接口
type Store interface {
    Save / GetByHash / GetByID / List / Update / Delete / AutoMigrate
}
```

**两种 Store 实现：**
- `NewGORMStore(db *gorm.DB) *GORMStore` — 生产用，基于 GORM（PostgreSQL/MySQL 等）
- `NewMemoryStore() *MemoryStore` — 测试用，基于内存，线程安全

**Manager 选项：**
- `WithRateLimiter(rl RateLimiter)` — 设置限流器（实现 `Allow(ctx, key string, limit int) (bool, error)`）
- `WithKeyPrefix(prefix string)` — 生成密钥的前缀（默认 `"sk-"`）

**创建选项（`CreateOption`）：**
- `WithName`, `WithOwnerID`, `WithPermissions`, `WithRateLimit`, `WithQuotaLimit`, `WithExpiresAt`

**HTTP 中间件：**
- `HTTPMiddleware(mgr Manager) func(http.Handler) http.Handler` — 从 `Authorization: Bearer sk-xxx` 或 `X-API-Key: sk-xxx` 提取并验证密钥，将 `*Key` 注入 context

**Context 辅助：**
- `FromContext(ctx) (*Key, bool)` — 从 context 中取出已验证的 Key
- `NewContext(ctx, key) context.Context` — 将 Key 注入 context

**错误常量：** `ErrKeyNotFound`、`ErrKeyDisabled`、`ErrKeyExpired`、`ErrQuotaExceeded`、`ErrRateLimited`、`ErrMissingKey`、`ErrInvalidKey`

```go
// 初始化 Manager（生产）
db := rdbms.MustNewDatabase(cfg, log).AsGORM()
store := apikey.NewGORMStore(db)
_ = store.AutoMigrate(ctx)

mgr, err := apikey.NewManager(store,
    apikey.WithKeyPrefix("sk-myapp-"),
)

// 签发密钥
rawKey, key, err := mgr.Create(ctx,
    apikey.WithName("production-key"),
    apikey.WithOwnerID("user-123"),
    apikey.WithQuotaLimit(1_000_000), // 100 万 token
    apikey.WithRateLimit(60),          // 60 RPM
    apikey.WithExpiresAt(time.Now().Add(365*24*time.Hour)),
)
fmt.Println("保存此密钥（仅显示一次）:", rawKey)

// 验证密钥
key, err = mgr.Validate(ctx, rawKey)
if errors.Is(err, apikey.ErrQuotaExceeded) {
    // 处理配额超限
}

// HTTP 中间件（与 servex httpserver 配合）
mux := http.NewServeMux()
mux.Handle("/v1/", apikey.HTTPMiddleware(mgr)(apiHandler))

// 在 handler 中获取已验证的 Key
func apiHandler(w http.ResponseWriter, r *http.Request) {
    key, _ := apikey.FromContext(r.Context())
    fmt.Println("调用方:", key.OwnerID)
}

// 更新配额
_ = mgr.UpdateQuota(ctx, key.ID, int64(resp.Usage.TotalTokens))
```

## llm/serving/billing — 用量计费

**核心类型：**

```go
// 定价模型
type PriceModel struct {
    ModelID         string
    InputPricePerM  float64 // 每 100 万输入 token 的价格
    OutputPricePerM float64 // 每 100 万输出 token 的价格
    CachedPricePerM float64 // 每 100 万缓存命中 token 的价格
}

// 用量记录（GORM 表：usage_records）
type UsageRecord struct {
    ID        string
    KeyID     string
    ModelID   string
    Usage     llm.Usage
    Cost      float64
    CreatedAt time.Time
}

// Billing 接口
type Billing interface {
    Record(ctx, keyID, modelID string, usage llm.Usage) error
    GetSummary(ctx, keyID string, from, to time.Time) (*Summary, error)
    SetPricing(modelID string, pricing PriceModel)
    CalculateCost(modelID string, usage llm.Usage) float64
}

// 汇总结果
type Summary struct {
    TotalRequests int64
    TotalTokens   int64
    TotalCost     float64
    ByModel       map[string]ModelSummary
}
```

**两种 Store 实现：**
- `NewGORMStore(db *gorm.DB) Store` — 生产用
- `NewMemoryStore() Store` — 测试用

**构造与选项：**
- `NewBilling(store Store, opts ...Option) Billing`
- `WithDefaultPricing(models []PriceModel)` — 初始化定价列表

**Middleware（与 `ai/middleware` 链配合）：**
- `Middleware(b Billing, keyExtractor func(ctx) string) llmmw.Middleware` — 在 Generate 响应后或流结束后异步记录用量（不阻塞主流程）

```go
// 初始化计费引擎
store := billing.NewGORMStore(db)
_ = store.AutoMigrate(ctx)

b := billing.NewBilling(store,
    billing.WithDefaultPricing([]billing.PriceModel{
        {ModelID: "gpt-4o",        InputPricePerM: 2.5,  OutputPricePerM: 10.0},
        {ModelID: "gpt-4o-mini",   InputPricePerM: 0.15, OutputPricePerM: 0.6},
        {ModelID: "claude-3-5-sonnet-20241022", InputPricePerM: 3.0, OutputPricePerM: 15.0},
    }),
)

// 手动记录
err = b.Record(ctx, "key-id-123", "gpt-4o", resp.Usage)

// 查询汇总
summary, err := b.GetSummary(ctx, "key-id-123",
    time.Now().AddDate(0, -1, 0), time.Now(),
)
fmt.Printf("本月请求：%d，总 Token：%d，费用：$%.4f\n",
    summary.TotalRequests, summary.TotalTokens, summary.TotalCost)

// 计算单次费用
cost := b.CalculateCost("gpt-4o", resp.Usage)

// 动态更新定价
b.SetPricing("gpt-4o", billing.PriceModel{
    ModelID: "gpt-4o", InputPricePerM: 2.0, OutputPricePerM: 8.0,
})

// 配合 ai/middleware 链自动计费
keyExtractor := func(ctx context.Context) string {
    if key, ok := apikey.FromContext(ctx); ok {
        return key.ID
    }
    return ""
}
billedModel := middleware.Chain(
    billing.Middleware(b, keyExtractor),
)(baseModel)
```

## llm/serving/proxy — AI API 代理网关

**核心类型：**

```go
// Proxy 结构（通过 New 构造）
type Proxy struct { ... }

// ProviderConfig（可用于配置文件反序列化）
type ProviderConfig struct {
    Name     string   // Provider 名称
    Models   []string // 支持的模型名列表
    Weight   int      // 负载均衡权重
    Priority int      // 故障转移优先级（越小越高）
}
```

**构造与 Provider 注册：**
- `New(providers map[string]llm.ChatModel, opts ...Option) *Proxy` — 创建 Proxy，可传入初始 providers map
- `(p *Proxy) RegisterProvider(name string, model llm.ChatModel, models []string, opts ...ProviderOption)` — 注册 Provider 并绑定支持的模型名；同名 Provider 会更新，模型名冲突时后注册者覆盖

**Provider 选项：**
- `WithWeight(w int)` — 负载均衡权重
- `WithPriority(p int)` — 故障转移优先级

**Proxy 构造选项：**
- `WithAPIKeyManager(mgr apikey.Manager)` — API Key 鉴权（需在 Handler 前挂载 `apikey.HTTPMiddleware`）
- `WithBilling(b billing.Billing)` — 自动记录 token 用量计费
- `WithModeration(mod moderation.Moderator)` — 请求内容审核
- `WithLogger(log logger.Logger)` — 日志记录

**路由与 HTTP Handler：**
- `(p *Proxy) Route(model string) (llm.ChatModel, error)` — 按模型名路由到对应 Provider
- `(p *Proxy) Handler() http.Handler` — 返回 OpenAI 兼容的 HTTP handler，注册以下路由：
  - `POST /v1/chat/completions` — 聊天补全（支持流式 SSE）
  - `GET /v1/models` — 模型列表

**错误常量：** `ErrModelNotFound`、`ErrNoProviders`、`ErrAllProvidersFailed`

```go
// 初始化各 Provider
openaiClient := openai.New(os.Getenv("OPENAI_API_KEY"), openllm.WithModel("gpt-4o"))
deepseekClient := openai.New(os.Getenv("DEEPSEEK_API_KEY"),
    openai.WithBaseURL("https://api.deepseek.com/v1"),
    openllm.WithModel("deepseek-chat"),
)
claudeClient := anthropic.New(os.Getenv("ANTHROPIC_API_KEY"),
    anthropic.WithModel("claude-3-5-sonnet-20241022"),
)

// 创建 Proxy（带鉴权、计费、审核）
p := proxy.New(nil,
    proxy.WithAPIKeyManager(mgr),
    proxy.WithBilling(b),
    proxy.WithModeration(kwMod),
    proxy.WithLogger(log),
)

// 注册 Provider 并绑定模型名
p.RegisterProvider("openai", openaiClient,
    []string{"gpt-4o", "gpt-4o-mini"},
    proxy.WithWeight(2),
)
p.RegisterProvider("deepseek", deepseekClient,
    []string{"deepseek-chat"},
    proxy.WithPriority(1),
)
p.RegisterProvider("anthropic", claudeClient,
    []string{"claude-3-5-sonnet-20241022"},
)

// 挂载到 HTTP 服务器
// apikey.HTTPMiddleware 负责验证并将 Key 注入 context，供 Proxy 内部使用
handler := apikey.HTTPMiddleware(mgr)(p.Handler())
http.ListenAndServe(":8080", handler)

// 客户端调用（标准 OpenAI 格式）
// POST http://localhost:8080/v1/chat/completions
// Authorization: Bearer sk-myapp-xxxxxxxx
// {"model": "gpt-4o", "messages": [...], "stream": false}

// 仅使用路由功能（不走 HTTP）
model, err := p.Route("gpt-4o")
if err != nil {
    // ErrModelNotFound / ErrNoProviders
}
resp, err := model.Generate(ctx, messages)
```

## llm/processing/classifier — 文本分类器

**Classifier 接口：**

```go
type Classifier interface {
    Classify(ctx context.Context, text string) (*Result, error)
    ClassifyMessages(ctx context.Context, messages []llm.Message) (*Result, error)
}

type Label struct {
    Name        string  `json:"name"`
    Score       float64 `json:"score"`        // 置信度 0.0-1.0
    Description string  `json:"description"`  // 判断理由
}

type Result struct {
    Labels []Label `json:"labels"` // 按 Score 降序排列
    Best   Label   `json:"best"`   // 最高分标签
}
```

**6 种内置分类器：**
- `NewIntentClassifier(model, intents map[string]string, opts...)` — 意图识别，`intents` 为意图名→描述映射
- `NewSentimentClassifier(model, opts...)` — 情感分析（positive/neutral/negative）
- `NewTopicClassifier(model, topics []string, opts...)` — 主题分类，`topics` 为空则 LLM 自动判断
- `NewLanguageClassifier(model, opts...)` — 语言检测，返回语言代码（zh/en/ja 等）
- `NewToxicityClassifier(model, opts...)` — 毒性检测（toxic/safe）
- `NewRouterClassifier(model, routes map[string]string, opts...)` — Agent/工具路由，按输入选择最匹配路由
- `NewCustomClassifier(model, labels []string, systemPrompt, opts...)` — 自定义标签与提示词

**选项：**
- `WithTopN(n)` — 只返回前 N 个标签（默认全部）
- `WithCallOptions(opts...)` — 透传底层模型调用选项

**错误常量：** `ErrNilModel`、`ErrEmptyText`、`ErrNoLabels`

```go
// 意图识别
intent := classifier.NewIntentClassifier(chatModel, map[string]string{
    "greeting":   "用户在打招呼",
    "question":   "用户在提问",
    "complaint":  "用户在投诉",
    "purchase":   "用户想购买商品",
}, classifier.WithTopN(2))

result, err := intent.Classify(ctx, "你好，我想买一台笔记本电脑")
if err != nil { ... }
fmt.Printf("最佳意图: %s (%.2f)\n", result.Best.Name, result.Best.Score)
// 最佳意图: purchase (0.92)

// 情感分析
sentiment := classifier.NewSentimentClassifier(chatModel)
result, err = sentiment.Classify(ctx, "这个产品质量真的太差了！")
fmt.Println(result.Best.Name) // negative

// 主题分类（固定主题列表）
topic := classifier.NewTopicClassifier(chatModel, []string{"科技", "体育", "财经", "娱乐"})
result, err = topic.Classify(ctx, "苹果发布 M4 芯片，AI 性能大幅提升")
fmt.Println(result.Best.Name) // 科技

// 语言检测
lang := classifier.NewLanguageClassifier(chatModel)
result, err = lang.Classify(ctx, "Hello, how are you?")
fmt.Println(result.Best.Name) // en

// 路由分类器（用于 Agent 分流）
router := classifier.NewRouterClassifier(chatModel, map[string]string{
    "search_agent":   "需要搜索互联网信息",
    "code_agent":     "需要编写或分析代码",
    "data_agent":     "需要处理数据或生成报表",
})
result, err = router.Classify(ctx, "帮我查一下最新的 Go 版本")
fmt.Printf("路由到: %s\n", result.Best.Name) // search_agent
```

## llm/processing/extractor — 信息提取器

**Extractor 接口：**

```go
type Extractor interface {
    Extract(ctx context.Context, text string) (*Result, error)
}

type Result struct {
    Entities  []Entity   `json:"entities,omitempty"`
    Relations []Relation `json:"relations,omitempty"`
    Keywords  []Keyword  `json:"keywords,omitempty"`
    Summary   *Summary   `json:"summary,omitempty"`
}

type Entity struct {
    Text     string         `json:"text"`
    Type     string         `json:"type"`  // person/organization/location/date 等
    Start    int            `json:"start"` // 原文起始位置，-1 表示未知
    End      int            `json:"end"`
    Metadata map[string]any `json:"metadata,omitempty"`
}

type Relation struct {
    Subject   string `json:"subject"`   // 主语实体
    Predicate string `json:"predicate"` // 关系谓词
    Object    string `json:"object"`    // 宾语实体
}

type Keyword struct {
    Word  string  `json:"word"`
    Score float64 `json:"score"` // 重要度 0.0-1.0
}

type Summary struct {
    Text      string `json:"text"`
    Sentences int    `json:"sentences"`
}
```

**4 种提取器：**
- `NewEntityExtractor(model, entityTypes []string, opts...)` — 实体识别，`entityTypes` 如 `["person", "organization", "location", "date"]`
- `NewRelationExtractor(model, opts...)` — 关系抽取，输出 subject→predicate→object 三元组
- `NewKeywordExtractor(model, opts...)` — 关键词提取，按重要度降序，支持 `WithMaxKeywords(n)`
- `NewSummarizer(model, opts...)` — 文本摘要，支持 `WithMaxSentences(n)`、`WithLanguage(lang)`

**关键词提取专属选项：**
- `WithMaxKeywords(n)` — 最大关键词数量（默认 10）

**摘要专属选项：**
- `WithMaxSentences(n)` — 摘要最大句子数（默认 3）
- `WithLanguage(lang)` — 摘要输出语言（如 `"en"`，默认与输入语言相同）

**通用选项：**
- `WithCallOptions(opts...)` — 透传底层模型调用选项

**错误常量：** `ErrNilModel`、`ErrEmptyText`

```go
// 实体识别
entityExt := extractor.NewEntityExtractor(chatModel,
    []string{"person", "organization", "location", "date"},
)
result, err := entityExt.Extract(ctx, "张三在 2024 年加入了北京的字节跳动公司")
if err != nil { ... }
for _, e := range result.Entities {
    fmt.Printf("[%s] %s\n", e.Type, e.Text)
}
// [person] 张三
// [date] 2024 年
// [location] 北京
// [organization] 字节跳动

// 关系抽取
relExt := extractor.NewRelationExtractor(chatModel)
result, err = relExt.Extract(ctx, "张三供职于字节跳动，李四是字节跳动的 CEO")
for _, r := range result.Relations {
    fmt.Printf("%s —[%s]→ %s\n", r.Subject, r.Predicate, r.Object)
}
// 张三 —[供职于]→ 字节跳动
// 李四 —[是CEO]→ 字节跳动

// 关键词提取
kwExt := extractor.NewKeywordExtractor(chatModel,
    extractor.WithMaxKeywords(5),
)
result, err = kwExt.Extract(ctx, longArticleText)
for _, kw := range result.Keywords {
    fmt.Printf("%s (%.2f)\n", kw.Word, kw.Score)
}

// 文本摘要（中文输入，英文摘要）
sum := extractor.NewSummarizer(chatModel,
    extractor.WithMaxSentences(2),
    extractor.WithLanguage("en"),
)
result, err = sum.Extract(ctx, longChineseText)
fmt.Println(result.Summary.Text)
```

## llm/processing/translator — 翻译器

**Translator 接口：**

```go
type Translator interface {
    Translate(ctx context.Context, text string, targetLang string) (*Translation, error)
    TranslateBatch(ctx context.Context, texts []string, targetLang string) (*BatchTranslation, error)
    DetectLanguage(ctx context.Context, text string) (string, error)
}

type Translation struct {
    Text           string `json:"text"`            // 翻译结果
    SourceLanguage string `json:"source_language"` // 源语言代码
    TargetLanguage string `json:"target_language"` // 目标语言代码
}

type BatchTranslation struct {
    Translations []Translation `json:"translations"` // 与输入文本一一对应
}
```

**构造器：**
- `NewTranslator(model, opts...)` — 创建 LLM 翻译器

**选项：**
- `WithSourceLanguage(lang)` — 指定源语言（默认自动检测）
- `WithGlossary(map[string]string)` — 术语表，确保专业词汇翻译一致性（key 源文，value 译文）
- `WithTone(tone)` — 翻译风格：`"formal"`（正式）/ `"informal"`（口语）/ `"technical"`（技术）
- `WithBatchSize(n)` — 批量翻译每批最大文本数（默认 10）
- `WithCallOptions(opts...)` — 透传底层模型调用选项

**错误常量：** `ErrNilModel`、`ErrEmptyText`、`ErrEmptyTarget`

```go
// 基础翻译（中译英）
tr := translator.NewTranslator(chatModel)
result, err := tr.Translate(ctx, "人工智能正在改变世界", "en")
if err != nil { ... }
fmt.Println(result.Text)           // Artificial intelligence is changing the world
fmt.Println(result.SourceLanguage) // zh

// 带术语表的技术翻译
techTr := translator.NewTranslator(chatModel,
    translator.WithTone("technical"),
    translator.WithGlossary(map[string]string{
        "微服务": "microservice",
        "熔断器": "circuit breaker",
        "服务网格": "service mesh",
    }),
)
result, err = techTr.Translate(ctx,
    "在微服务架构中，熔断器是保护系统稳定性的关键组件", "en")
fmt.Println(result.Text)
// In microservice architecture, circuit breaker is a key component for protecting system stability

// 批量翻译
batch, err := tr.TranslateBatch(ctx, []string{
    "你好世界",
    "Go 语言很棒",
    "欢迎使用 servex",
}, "en")
for i, t := range batch.Translations {
    fmt.Printf("[%d] %s\n", i, t.Text)
}

// 语言检测
lang, err := tr.DetectLanguage(ctx, "Bonjour le monde")
fmt.Println(lang) // fr
```

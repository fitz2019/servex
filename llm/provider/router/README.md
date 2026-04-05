# ai/router

`ai/router` 包提供多 Provider 路由器，实现 `llm.ChatModel` 接口，根据调用时的 `WithModel()` 选项将请求转发到对应的 Provider 客户端。

## 功能特性

- 按模型名精确匹配路由，第一个命中的路由生效
- 无匹配（或未指定模型）时自动走 fallback
- 完整实现 `llm.ChatModel`，包括 `Generate` 和 `Stream`
- 与 `ai/middleware` 无缝组合

## 安装

```bash
go get github.com/Tsukikage7/servex/llm
```

## API

```go
type Route struct {
    Models []string     // 此路由支持的模型名列表（精确匹配）
    Model  llm.ChatModel // 对应的 Provider 客户端
}

func New(fallback llm.ChatModel, routes ...Route) *Router

func (r *Router) Generate(ctx context.Context, messages []llm.Message, opts ...llm.CallOption) (*llm.ChatResponse, error)
func (r *Router) Stream(ctx context.Context, messages []llm.Message, opts ...llm.CallOption) (llm.StreamReader, error)
```

**路由选择逻辑：**
1. 从 `opts` 中提取 `WithModel()` 指定的模型名
2. 若为空 → fallback
3. 遍历 `routes`，返回第一个 `Models` 包含该名称的条目
4. 无命中 → fallback

## 使用示例

```go
// 构建各 Provider 客户端
openaiClient := openllm.New(os.Getenv("OPENAI_API_KEY"),
    openllm.WithModel("gpt-4o"),
)
dashscopeClient := openllm.New(os.Getenv("DASHSCOPE_API_KEY"),
    openllm.WithBaseURL("https://dashscope.aliyuncs.com/compatible-mode/v1"),
)
claudeClient := anthropic.New(os.Getenv("ANTHROPIC_API_KEY"))

// 构建路由器
r := router.New(
    openaiClient, // fallback：未匹配时使用 OpenAI
    router.Route{
        Models: []string{"qwen-plus", "qwen-max", "qwen-turbo"},
        Model:  dashscopeClient,
    },
    router.Route{
        Models: []string{"claude-opus-4-6", "claude-sonnet-4-6"},
        Model:  claudeClient,
    },
)

// 路由到 DashScope
resp, _ := r.Generate(ctx, messages, llm.WithModel("qwen-plus"))

// 路由到 Anthropic
resp, _ = r.Generate(ctx, messages, llm.WithModel("claude-opus-4-6"))

// 走 fallback（OpenAI）
resp, _ = r.Generate(ctx, messages)
```

### 与中间件组合

```go
// 路由器本身是 llm.ChatModel，可直接套中间件
chain := aimw.Chain(
    aimw.Retry(3, 500*time.Millisecond),
    aimw.Logging(log),
)
model := chain(r) // r 是 *router.Router
```

## 许可证

详见项目根目录 LICENSE 文件。

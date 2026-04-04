# ai

`ai` 包提供 LLM/AI 服务的统一客户端抽象，支持多 Provider（OpenAI、Anthropic、Gemini 等），所有适配器仅依赖标准库。

## 功能特性

- 统一 `ChatModel` / `EmbeddingModel` 接口，Provider 可随时替换
- 流式与非流式生成，支持工具调用（Function Calling）
- 多模态消息（文本 + 图片）
- 统一错误类型与可重试判断
- 子包覆盖：Provider 适配、中间件链、多轮会话、工具调用、提示词模板、嵌入向量、向量存储接口、多 Provider 路由

## 安装

```bash
go get github.com/Tsukikage7/servex/ai
```

## 子包总览

| 子包 | 说明 |
|---|---|
| `ai/openai` | OpenAI 适配器（兼容 DeepSeek、通义千问等 OpenAI 格式 Provider） |
| `ai/anthropic` | Anthropic Claude 适配器 |
| `ai/gemini` | Google Gemini 适配器 |
| `ai/middleware` | 中间件链（日志、重试、限流、用量追踪） |
| `ai/conversation` | 多轮对话会话管理（BufferMemory / WindowMemory） |
| `ai/toolcall` | 工具注册与自动循环执行器（ReAct 模式） |
| `ai/prompt` | 基于 `text/template` 的提示词模板引擎 |
| `ai/embedding` | 批量嵌入 + 余弦相似度工具函数 |
| `ai/vectorstore` | 向量存储统一接口抽象 |
| `ai/router` | 多 Provider 路由器（按模型名路由） |

## 核心接口

```go
// 聊天模型
type ChatModel interface {
    Generate(ctx context.Context, messages []Message, opts ...CallOption) (*ChatResponse, error)
    Stream(ctx context.Context, messages []Message, opts ...CallOption) (StreamReader, error)
}

// 嵌入模型
type EmbeddingModel interface {
    EmbedTexts(ctx context.Context, texts []string, opts ...CallOption) (*EmbedResponse, error)
}
```

## 消息构造辅助函数

```go
ai.SystemMessage("你是一个专业助手")
ai.UserMessage("帮我写一首诗")
ai.AssistantMessage("好的，...")
ai.ToolResultMessage(callID, `{"result": "ok"}`)
```

## 调用选项

```go
ai.WithModel("gpt-4o")
ai.WithTemperature(0.7)
ai.WithMaxTokens(1024)
ai.WithTopP(0.9)
ai.WithStop("END")
ai.WithTools(tool1, tool2)
ai.WithToolChoice(ai.ToolChoiceAuto)
ai.WithStreamCallback(fn)
```

## 错误处理

```go
// 哨兵错误
ai.ErrRateLimited          // HTTP 429
ai.ErrContextLength        // 上下文超长
ai.ErrInvalidAuth          // 认证失败
ai.ErrProviderUnavailable  // 服务不可用（5xx）
ai.ErrContentFiltered      // 内容过滤

// 可重试判断
if ai.IsRetryable(err) { ... }

// 获取详细错误信息
var apiErr *ai.APIError
if errors.As(err, &apiErr) {
    fmt.Println(apiErr.StatusCode, apiErr.Provider, apiErr.Message)
}
```

## 快速示例

```go
client := openai.New(os.Getenv("OPENAI_API_KEY"),
    openai.WithModel("gpt-4o"),
)

resp, err := client.Generate(ctx, []ai.Message{
    ai.UserMessage("你好"),
})
fmt.Println(resp.Message.Content)
```

## 许可证

详见项目根目录 LICENSE 文件。

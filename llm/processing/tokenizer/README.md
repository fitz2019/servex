# llm/processing/tokenizer

`github.com/Tsukikage7/servex/llm/processing/tokenizer` — LLM Token 计数与截断工具，用于成本控制和上下文窗口管理。

## 核心类型

- `Tokenizer` — Token 计数器接口，方法包括 `Count(text)`、`CountMessages(messages)`、`Truncate(text, maxTokens)`
- `NewEstimateTokenizer(opts...)` — 基于字符比例的估算 Tokenizer，可自定义每 token 字符数
- `NewCL100KTokenizer()` — GPT-4/Claude 系列规则估算 Tokenizer（英文 4 chars/token，中文 1.5 chars/token）
- `EstimateTokens(text)` — 包级快捷函数，使用默认 CL100K Tokenizer 估算 Token 数
- `EstimateMessageTokens(messages)` — 估算消息列表的总 Token 数（含每条消息固定开销）
- `FitsContext(messages, maxTokens)` — 判断消息列表是否在上下文窗口内
- `TruncateToFit(messages, maxTokens)` — 截断消息列表以满足 Token 限制（保留系统消息，丢弃最旧非系统消息）

## 使用示例

```go
import "github.com/Tsukikage7/servex/llm/processing/tokenizer"

// 快捷函数
n := tokenizer.EstimateTokens("你好，世界！Hello, world!")
fmt.Printf("约 %d tokens\n", n)

// 判断是否超出上下文
if !tokenizer.FitsContext(messages, 4096) {
    messages = tokenizer.TruncateToFit(messages, 4096)
}

// 自定义 Tokenizer
tk := tokenizer.NewEstimateTokenizer(
    tokenizer.WithCharsPerToken(4.0),
    tokenizer.WithCJKCharsPerToken(1.5),
)
count := tk.Count("人工智能")
truncated := tk.Truncate("很长的文本...", 100)
_ = truncated
```

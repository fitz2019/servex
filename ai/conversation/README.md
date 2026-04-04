# ai/conversation

`ai/conversation` 包提供多轮对话会话管理，自动维护消息历史并注入系统提示词。

## 功能特性

- 自动管理对话历史，每次 `Chat` 时将历史 + 当前输入一起发送
- 支持系统提示词（`WithSystemPrompt`）
- 可插拔记忆策略：`BufferMemory`（完整历史）/ `WindowMemory`（滑动窗口）
- 流式对话（`ChatStream`）自动在流结束后将助手消息写入记忆
- 失败时不污染历史（`Chat` 失败不保留用户消息）

## 安装

```bash
go get github.com/Tsukikage7/servex/ai
```

## API

### Conversation

```go
func New(model ai.ChatModel, opts ...Option) *Conversation

func (c *Conversation) Chat(ctx context.Context, input string, opts ...ai.CallOption) (*ai.ChatResponse, error)
func (c *Conversation) ChatStream(ctx context.Context, input string, opts ...ai.CallOption) (ai.StreamReader, error)
func (c *Conversation) History() []ai.Message
func (c *Conversation) Reset()
```

### 选项

| 选项 | 说明 |
|---|---|
| `WithMemory(m Memory)` | 设置记忆策略（默认 `BufferMemory`） |
| `WithSystemPrompt(prompt)` | 设置系统提示词，每次请求自动注入 |

### 记忆策略

```go
// 完整历史（默认）
memory := conversation.NewBufferMemory()

// 滑动窗口：只保留最近 N 轮（每轮 = 用户 + 助手各一条）
memory := conversation.NewWindowMemory(10)
```

## 使用示例

```go
client := openai.New(apiKey)

conv := conversation.New(client,
    conversation.WithSystemPrompt("你是一个专业的 Go 语言助手"),
    conversation.WithMemory(conversation.NewWindowMemory(20)),
)

// 第一轮
resp, _ := conv.Chat(ctx, "什么是 goroutine？")
fmt.Println(resp.Message.Content)

// 第二轮（自动携带第一轮历史）
resp, _ = conv.Chat(ctx, "和线程有什么区别？")
fmt.Println(resp.Message.Content)

// 查看历史
for _, msg := range conv.History() {
    fmt.Printf("[%s] %s\n", msg.Role, msg.Content)
}

// 重置会话
conv.Reset()
```

## 许可证

详见项目根目录 LICENSE 文件。

# ai/toolcall

`ai/toolcall` 包提供工具调用（Function Calling）框架，支持工具注册和自动 ReAct 循环执行。

## 功能特性

- `Registry`：工具注册表，管理工具定义与处理函数
- `Executor`：自动循环执行 `model → tool_calls → execute → result → model`，直到模型停止调用工具或达到最大轮次
- `WithOnStep` 回调：每轮处理完毕后触发，可用于 SSE 实时推送思考链、Redis 步骤存储等

## 安装

```bash
go get github.com/Tsukikage7/servex/ai
```

## API

### Registry

```go
func NewRegistry() *Registry

func (r *Registry) Register(tool ai.Tool, handler HandlerFunc)
func (r *Registry) Tools() []ai.Tool
func (r *Registry) Execute(ctx context.Context, callID, name, arguments string) (string, error)

// 工具处理函数签名
type HandlerFunc func(ctx context.Context, arguments string) (string, error)
```

### Executor

```go
func NewExecutor(model ai.ChatModel, registry *Registry, opts ...ExecutorOption) *Executor
func (e *Executor) Run(ctx context.Context, messages []ai.Message, opts ...ai.CallOption) (*ExecutorResult, error)
```

### ExecutorResult

```go
type ExecutorResult struct {
    Response *ai.ChatResponse // 最终响应（工具调用循环结束后）
    Messages []ai.Message     // 完整对话历史（含所有工具调用和结果）
    Rounds   int              // 实际执行的工具调用轮次数
}
```

### 执行器选项

| 选项 | 默认值 | 说明 |
|---|---|---|
| `WithMaxRounds(n)` | `10` | 最大工具调用轮次 |
| `WithOnStep(fn)` | - | 每轮推理步骤回调 |

### StepEvent（步骤回调）

```go
type StepEvent struct {
    Round       int              // 轮次序号（0-based）
    Response    *ai.ChatResponse // 本轮 LLM 响应
    ToolResults []ToolResult     // 本轮工具执行结果（IsFinal=true 时为 nil）
    IsFinal     bool             // 是否为最终轮（模型不再调用工具）
}

type ToolResult struct {
    Call   ai.ToolCall // 原始工具调用请求
    Output string      // 执行结果（JSON 字符串）
    Err    error       // 执行错误（nil 表示成功）
}

type StepHandler func(ctx context.Context, event StepEvent)
```

## 使用示例

### 基础用法

```go
reg := toolcall.NewRegistry()

// 注册工具
reg.Register(
    ai.Tool{
        Function: ai.FunctionDef{
            Name:        "get_weather",
            Description: "获取指定城市的天气",
            Parameters:  json.RawMessage(`{
                "type": "object",
                "properties": {
                    "city": {"type": "string", "description": "城市名称"}
                },
                "required": ["city"]
            }`),
        },
    },
    func(ctx context.Context, args string) (string, error) {
        var params struct{ City string }
        json.Unmarshal([]byte(args), &params)
        // 实际调用天气 API...
        return `{"temperature": 25, "condition": "晴"}`, nil
    },
)

executor := toolcall.NewExecutor(client, reg,
    toolcall.WithMaxRounds(5),
)

result, err := executor.Run(ctx, []ai.Message{
    ai.UserMessage("北京今天天气怎么样？"),
})
fmt.Println(result.Response.Message.Content)
fmt.Printf("执行了 %d 轮工具调用\n", result.Rounds)
```

### 带步骤回调（实时推送 SSE）

```go
executor := toolcall.NewExecutor(client, reg,
    toolcall.WithOnStep(func(ctx context.Context, event toolcall.StepEvent) {
        if event.IsFinal {
            // 最终回复
            sseWriter.Write("final", event.Response.Message.Content)
            return
        }
        // 工具调用轮：推送思考过程
        for _, tr := range event.ToolResults {
            sseWriter.Write("tool", fmt.Sprintf("调用 %s: %s",
                tr.Call.Function.Name, tr.Output))
        }
    }),
)
```

## 许可证

详见项目根目录 LICENSE 文件。

# llm/agent

`github.com/Tsukikage7/servex/llm/agent` — AI Agent 框架，支持 ReAct/PlanExecute 执行策略、护栏、记忆管理及多 Agent 编排。

## 核心类型

- `Agent` — 智能代理主体，封装模型调用、工具执行、记忆与护栏
- `Config` — Agent 配置，包含 Model、SystemPrompt、Tools、Memory、Guardrails、Strategy 等字段
- `Strategy` — 执行策略接口，内置 `ReActStrategy`（思考-行动循环）和 `PlanExecuteStrategy`（计划-执行）
- `Result` — 执行结果，包含 Output、Messages、ToolCalls、Iterations、Usage
- `Event` — 流式事件，类型包括 `thinking`、`tool_call`、`tool_result`、`output`、`error`
- `Supervisor` — 监督者模式，将多个 Worker Agent 注册为工具并由监督者调度
- `Pipeline` — 管道模式，多个 Agent 串行执行，前一个输出作为后一个输入

## 使用示例

```go
import "github.com/Tsukikage7/servex/llm/agent"

a, err := agent.New(&agent.Config{
    Name:          "assistant",
    Model:         myModel,
    SystemPrompt:  "你是一个助手",
    MaxIterations: 10,
    // Strategy 默认为 ReAct
})
if err != nil {
    log.Fatal(err)
}

result, err := a.Run(ctx, "帮我查询北京今天的天气")
if err != nil {
    log.Fatal(err)
}
fmt.Println(result.Output)

// 流式执行
ch, _ := a.RunStream(ctx, "写一首诗")
for evt := range ch {
    if evt.Type == agent.EventOutput {
        fmt.Println(evt.Content)
    }
}

// 多 Agent：管道模式
pipeline := agent.NewPipeline(summarizerAgent, translatorAgent)
pipeResult, _ := pipeline.Run(ctx, "原始文本...")
```

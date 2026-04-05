# llm/agent/chain

`github.com/Tsukikage7/servex/llm/agent/chain` — 多步 AI 编排链，将多个 prompt/模型调用串联顺序执行。

## 核心类型

- `Chain` — 编排链主体，按顺序执行所有步骤
- `Step` — 单个步骤，包含 Name、Prompt（模板）、Model（可选，覆盖默认）、Parser（可选，解析输出）
- `Result` — 链执行结果，包含最后一步的 Output、Parsed、各步骤的 Steps 列表及累计 Usage
- `StepResult` — 单步结果，含 Name、Output、Parsed、Duration、Usage
- `StepEvent` — 步骤完成事件，通过 `WithOnStep` 回调触发
- `WithModel(model)` — 设置链的默认模型
- `WithOnStep(fn)` — 设置步骤完成后的回调

## 使用示例

```go
import (
    "github.com/Tsukikage7/servex/llm/agent/chain"
    "github.com/Tsukikage7/servex/llm/prompt"
)

summaryTmpl, _ := prompt.New(llm.RoleUser, "请总结以下内容：{{.}}")
translateTmpl, _ := prompt.New(llm.RoleUser, "将以下内容翻译为英文：{{.Previous}}")

c := chain.New(chain.WithModel(myModel)).
    AddStep(chain.Step{Name: "summarize", Prompt: summaryTmpl}).
    AddStep(chain.Step{Name: "translate", Prompt: translateTmpl})

result, err := c.Run(ctx, "这是一段很长的中文文档...")
if err != nil {
    log.Fatal(err)
}
fmt.Println(result.Output)
```

# llm/safety/moderation

`github.com/Tsukikage7/servex/llm/safety/moderation` — 内容审核，按类别检测有害内容，支持 LLM 审核、关键词审核及组合审核。

## 核心类型

- `Moderator` — 审核器接口，方法包括 `Moderate(ctx, text)` 和 `ModerateMessages(ctx, messages)`
- `Result` — 审核结果，包含 Flagged（是否违规）、Categories（各类别命中）、Scores（各类别分数）、Reason
- `Category` — 审核类别，内置 `violence`、`sexual`、`hate`、`self_harm`、`dangerous`、`political`、`spam`
- `NewLLMModerator(model, opts...)` — 基于 LLM 的审核器，可配置分数阈值和检测类别子集
- `NewKeywordModerator(rules)` — 基于关键词匹配的快速审核器（rules 为 Category -> 关键词列表的映射）
- `NewCompositeModerator(...)` — 组合审核器，关键词先行短路，避免不必要的 LLM 调用
- `WithThreshold(t)` — 设置触发标记的分数阈值（默认 0.7）
- `WithCategories(cats...)` — 设置需检测的类别子集

## 使用示例

```go
import "github.com/Tsukikage7/servex/llm/safety/moderation"

// LLM 审核器
mod := moderation.NewLLMModerator(myModel,
    moderation.WithThreshold(0.8),
    moderation.WithCategories(moderation.CategoryViolence, moderation.CategoryHate),
)
result, err := mod.Moderate(ctx, "用户输入的文本")
if result.Flagged {
    fmt.Printf("内容违规: %s\n", result.Reason)
}

// 组合审核（关键词快速过滤 + LLM 深度审核）
composite := moderation.NewCompositeModerator(
    moderation.NewKeywordModerator(map[moderation.Category][]string{
        moderation.CategoryViolence: {"打架", "伤害"},
    }),
    moderation.NewLLMModerator(myModel),
)
result, _ = composite.Moderate(ctx, "需要审核的内容")
```

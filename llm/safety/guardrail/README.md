# llm/safety/guardrail

`github.com/Tsukikage7/servex/llm/safety/guardrail` — AI 输入/输出护栏，用于过滤有害内容、PII 信息及限制消息规模，可作为中间件集成到模型调用链。

## 核心类型

- `Guard` — 护栏接口，方法为 `Check(ctx, messages) error`
- `GuardFunc` — 函数形式的 Guard，可将任意函数包装为 Guard
- `MaxLength(maxChars)` — 限制所有消息总字符数，超限返回 `ErrTooLong`
- `MaxMessages(n)` — 限制消息数量，超限返回 `ErrTooMany`
- `KeywordFilter(keywords)` — 关键词过滤（大小写不敏感），命中返回 `ErrBlocked`
- `RegexFilter(patterns)` — 正则过滤，命中返回 `ErrBlocked`
- `PIIDetector(patterns...)` — PII 检测，内置 email/phone/id_card/credit_card，命中返回 `ErrPIIDetected`
- `Middleware(opts...)` — 返回护栏中间件，输入护栏在调用前执行，输出护栏在返回后执行
- `WithInputGuards(guards...)` — 设置输入护栏
- `WithOutputGuards(guards...)` — 设置输出护栏

## 使用示例

```go
import "github.com/Tsukikage7/servex/llm/safety/guardrail"

// 单独使用
g := guardrail.KeywordFilter([]string{"暴力", "违禁"})
if err := g.Check(ctx, messages); err != nil {
    log.Printf("被护栏拦截: %v", err)
}

// 作为中间件包装模型
safeMw := guardrail.Middleware(
    guardrail.WithInputGuards(
        guardrail.MaxLength(10000),
        guardrail.PIIDetector(guardrail.PIIEmail, guardrail.PIIPhone),
        guardrail.KeywordFilter([]string{"敏感词"}),
    ),
    guardrail.WithOutputGuards(
        guardrail.MaxLength(5000),
    ),
)
safeModel := safeMw(myModel)
resp, err := safeModel.Generate(ctx, messages)
```

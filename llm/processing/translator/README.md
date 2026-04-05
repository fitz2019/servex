# llm/processing/translator

`github.com/Tsukikage7/servex/llm/processing/translator` — 基于 LLM 的文本翻译，支持单文本翻译、批量翻译和语言检测。

## 核心类型

- `Translator` — 翻译器接口，方法包括 `Translate`、`TranslateBatch`、`DetectLanguage`
- `Translation` — 翻译结果，包含 Text、SourceLanguage、TargetLanguage
- `BatchTranslation` — 批量翻译结果，包含 `[]Translation`（与输入一一对应）
- `NewTranslator(model, opts...)` — 创建基于 LLM 的翻译器
- `WithSourceLanguage(lang)` — 指定源语言（默认自动检测）
- `WithGlossary(map)` — 设置术语表，确保专业词汇翻译一致
- `WithTone(tone)` — 设置翻译风格（formal/informal/technical）
- `WithBatchSize(n)` — 设置批量翻译每批大小（默认 10）

## 使用示例

```go
import "github.com/Tsukikage7/servex/llm/processing/translator"

t := translator.NewTranslator(myModel,
    translator.WithTone("formal"),
    translator.WithGlossary(map[string]string{
        "人工智能": "Artificial Intelligence",
    }),
)

// 单条翻译
result, err := t.Translate(ctx, "人工智能正在改变世界", "en")
fmt.Printf("[%s->%s] %s\n", result.SourceLanguage, result.TargetLanguage, result.Text)

// 批量翻译
batch, _ := t.TranslateBatch(ctx, []string{"你好", "谢谢", "再见"}, "en")
for _, tr := range batch.Translations {
    fmt.Println(tr.Text)
}

// 语言检测
lang, _ := t.DetectLanguage(ctx, "Bonjour le monde")
fmt.Println(lang) // fr
```

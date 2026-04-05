# llm/processing/extractor

`github.com/Tsukikage7/servex/llm/processing/extractor` — 基于 LLM 的信息提取，支持实体识别、关系抽取、关键词提取和文本摘要。

## 核心类型

- `Extractor` — 提取器接口，方法为 `Extract(ctx, text) (*Result, error)`
- `Result` — 提取结果，包含 Entities、Relations、Keywords、Summary（按提取器类型填充）
- `Entity` — 实体，包含 Text、Type、Start、End、Metadata
- `Relation` — 关系三元组，包含 Subject、Predicate、Object
- `Keyword` — 关键词，包含 Word、Score（重要度）
- `Summary` — 摘要，包含 Text、Sentences
- `NewEntityExtractor(model, entityTypes)` — 命名实体识别
- `NewRelationExtractor(model)` — 实体关系抽取
- `NewKeywordExtractor(model, opts...)` — 关键词提取，可用 `WithMaxKeywords(n)` 限制数量
- `NewSummarizer(model, opts...)` — 文本摘要，可用 `WithMaxSentences(n)` 和 `WithLanguage(lang)` 配置

## 使用示例

```go
import "github.com/Tsukikage7/servex/llm/processing/extractor"

// 实体识别
ner := extractor.NewEntityExtractor(myModel, []string{"person", "organization", "location"})
result, err := ner.Extract(ctx, "张三在北京字节跳动公司工作。")
for _, e := range result.Entities {
    fmt.Printf("[%s] %s\n", e.Type, e.Text)
}

// 文本摘要
sum := extractor.NewSummarizer(myModel,
    extractor.WithMaxSentences(2),
    extractor.WithLanguage("zh"),
)
sumResult, _ := sum.Extract(ctx, "这是一篇很长的文章...")
fmt.Println(sumResult.Summary.Text)

// 关键词
kw := extractor.NewKeywordExtractor(myModel, extractor.WithMaxKeywords(5))
kwResult, _ := kw.Extract(ctx, "人工智能正在改变我们的生活方式...")
for _, k := range kwResult.Keywords {
    fmt.Printf("%s (%.2f)\n", k.Word, k.Score)
}
```

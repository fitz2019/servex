# llm/processing/classifier

`github.com/Tsukikage7/servex/llm/processing/classifier` — 基于 LLM 的文本分类，支持意图识别、情感分析、主题分类、语言检测、毒性检测及自定义路由。

## 核心类型

- `Classifier` — 分类器接口，方法包括 `Classify(ctx, text)` 和 `ClassifyMessages(ctx, messages)`
- `Result` — 分类结果，包含 `Labels`（按置信度降序）和 `Best`（最高分标签）
- `Label` — 单个标签，包含 Name、Score（0.0-1.0）、Description
- `NewIntentClassifier(model, intents)` — 意图识别分类器，intents 为意图名到描述的映射
- `NewSentimentClassifier(model)` — 情感分析（positive/neutral/negative）
- `NewTopicClassifier(model, topics)` — 主题分类器，topics 为空则自动识别
- `NewLanguageClassifier(model)` — 语言检测分类器
- `NewToxicityClassifier(model)` — 毒性检测分类器
- `NewRouterClassifier(model, routes)` — Agent/工具路由分类器
- `NewCustomClassifier(model, labels, prompt)` — 自定义分类器

## 使用示例

```go
import "github.com/Tsukikage7/servex/llm/processing/classifier"

// 情感分析
c := classifier.NewSentimentClassifier(myModel)
result, err := c.Classify(ctx, "这个产品真的太好用了！")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("最佳标签: %s (%.2f)\n", result.Best.Name, result.Best.Score)

// 意图识别
intentC := classifier.NewIntentClassifier(myModel, map[string]string{
    "booking":  "预订机票或酒店",
    "query":    "查询信息",
    "cancel":   "取消订单",
})
intentResult, _ := intentC.Classify(ctx, "帮我订一张明天去上海的机票")
fmt.Println(intentResult.Best.Name) // booking
```

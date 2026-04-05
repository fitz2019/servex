# llm/eval

`github.com/Tsukikage7/servex/llm/eval` — LLM 输出质量评估框架，支持相关性、忠实性、连贯性、正确性等多维度评估。

## 核心类型

- `Evaluator` — 评估器接口，方法为 `Evaluate(ctx, EvalInput) (*EvalResult, error)`
- `EvalInput` — 评估输入，包含 Question、Answer、Reference（参考答案）、Context（参考资料列表）
- `EvalResult` — 评估结果，包含 `[]Score`（各维度评分）
- `Score` — 单项评分，包含 Name、Value（0.0-1.0）、Reason
- `RelevanceEvaluator(model)` — 创建相关性评估器
- `FaithfulnessEvaluator(model)` — 创建忠实性评估器（基于参考资料）
- `CoherenceEvaluator(model)` — 创建连贯性评估器
- `CorrectnessEvaluator(model)` — 创建正确性评估器（基于参考答案）
- `NewCompositeEvaluator(...)` — 创建组合评估器，并发运行所有子评估器并合并结果

## 使用示例

```go
import "github.com/Tsukikage7/servex/llm/eval"

// 单一维度评估
relevance := eval.RelevanceEvaluator(myModel)
result, err := relevance.Evaluate(ctx, eval.EvalInput{
    Question: "什么是机器学习？",
    Answer:   "机器学习是一种让计算机从数据中学习的技术。",
})
fmt.Printf("相关性: %.2f，理由：%s\n", result.Scores[0].Value, result.Scores[0].Reason)

// 多维度组合评估
composite := eval.NewCompositeEvaluator(
    eval.RelevanceEvaluator(myModel),
    eval.CoherenceEvaluator(myModel),
    eval.CorrectnessEvaluator(myModel),
)
multiResult, _ := composite.Evaluate(ctx, eval.EvalInput{
    Question:  "首都是哪里？",
    Answer:    "北京是中国的首都。",
    Reference: "中国的首都是北京。",
})
for _, s := range multiResult.Scores {
    fmt.Printf("%s: %.2f\n", s.Name, s.Value)
}
```

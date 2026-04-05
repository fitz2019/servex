# llm/retrieval/rerank

`github.com/Tsukikage7/servex/llm/retrieval/rerank` — RAG 检索结果重排序，支持基于 LLM 评分、Embedding 余弦相似度及外部 Cross-Encoder API 三种策略。

## 核心类型

- `Reranker` — 重排序接口，方法为 `Rerank(ctx, query, docs) ([]rag.RetrievedDoc, error)`
- `NewLLMReranker(model, opts...)` — 基于 LLM 为每篇文档打分后重排序，支持分批处理
- `NewEmbeddingReranker(model, opts...)` — 基于 Embedding 余弦相似度重排序
- `NewCrossEncoderReranker(endpoint, opts...)` — 调用外部 Cross-Encoder API 重排序
- `WithTopN(n)` — 返回前 N 篇文档
- `WithBatchSize(n)` — LLM 重排序时每批处理文档数
- `WithAPIKey(key)` — Cross-Encoder API 密钥
- `WithModel(model)` — Cross-Encoder 模型名称

## 使用示例

```go
import "github.com/Tsukikage7/servex/llm/retrieval/rerank"

// LLM 重排序
reranker := rerank.NewLLMReranker(myModel,
    rerank.WithTopN(3),
    rerank.WithBatchSize(10),
)

// 先通过 RAG 检索
docs, _ := pipeline.Retrieve(ctx, "什么是向量数据库？")

// 重排序
ranked, err := reranker.Rerank(ctx, "什么是向量数据库？", docs)
for i, d := range ranked {
    fmt.Printf("%d. [%.3f] %s\n", i+1, d.Score, d.ID)
}

// Embedding 重排序
embReranker := rerank.NewEmbeddingReranker(embModel, rerank.WithTopN(5))

// Cross-Encoder 重排序
ceReranker := rerank.NewCrossEncoderReranker(
    "https://api.example.com/rerank",
    rerank.WithAPIKey("sk-xxx"),
)
```

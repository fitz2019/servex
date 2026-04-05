# llm/retrieval/rag

`github.com/Tsukikage7/servex/llm/retrieval/rag` — 检索增强生成（RAG）管线，支持文档导入（分块、嵌入、存储）、语义检索及非流式/流式生成。

## 核心类型

- `Pipeline` — RAG 管线主体
- `Config` — 管线配置，必填 ChatModel、EmbeddingModel、VectorStore，可选 Splitter、TopK、ScoreThreshold、PromptTemplate
- `Document` — 待导入文档，包含 ID、Content、Metadata
- `RetrievedDoc` — 检索结果，嵌套 Document 并附加 Score（相似度）
- `Result` — RAG 结果，包含 Answer、Sources（检索文档列表）、Usage
- `New(cfg)` — 创建 RAG 管线，校验必填配置并填充默认值（TopK=5）

## 主要方法

- `Ingest(ctx, docs)` — 导入文档：分块（可选）→ 嵌入 → 存入向量库
- `Retrieve(ctx, question)` — 仅检索，返回相关文档列表
- `Query(ctx, question, opts...)` — 检索增强生成，返回 `*Result`
- `QueryStream(ctx, question, opts...)` — 流式检索增强生成，返回 `llm.StreamReader`

## 使用示例

```go
import "github.com/Tsukikage7/servex/llm/retrieval/rag"

pipeline, err := rag.New(&rag.Config{
    ChatModel:      chatModel,
    EmbeddingModel: embModel,
    VectorStore:    vs,
    TopK:           5,
})
if err != nil {
    log.Fatal(err)
}

// 导入文档
_ = pipeline.Ingest(ctx, []rag.Document{
    {ID: "doc1", Content: "Go 是一种静态类型、编译型语言..."},
})

// 检索增强生成
result, err := pipeline.Query(ctx, "Go 语言有什么特点？")
fmt.Println(result.Answer)
for _, s := range result.Sources {
    fmt.Printf("  来源: %s (score=%.3f)\n", s.ID, s.Score)
}
```

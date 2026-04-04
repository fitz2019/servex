# ai/vectorstore

`ai/vectorstore` 包提供向量存储的统一接口抽象，解耦业务代码与具体向量数据库实现。

## 功能特性

- `VectorStore` 接口：增删查，支持元数据过滤和相似度阈值
- `Document`：统一文档结构（ID、内容、向量、元数据）
- `SearchResult`：相似度搜索结果（文档 + 分数）
- 搜索选项：`WithFilter`（元数据过滤）、`WithScoreThreshold`（分数阈值）

## 安装

```bash
go get github.com/Tsukikage7/servex/ai
```

## 接口定义

```go
type VectorStore interface {
    AddDocuments(ctx context.Context, docs []Document) error
    SimilaritySearch(ctx context.Context, query []float32, k int, opts ...SearchOption) ([]SearchResult, error)
    Delete(ctx context.Context, ids []string) error
}

type Document struct {
    ID       string
    Content  string
    Vector   []float32
    Metadata map[string]any
}

type SearchResult struct {
    Document Document
    Score    float32
}
```

## 搜索选项

```go
vectorstore.WithFilter(map[string]any{"category": "tech"})  // 元数据过滤
vectorstore.WithScoreThreshold(0.8)                          // 仅返回相似度 ≥ 0.8 的结果
```

## 使用示例

RAG（检索增强生成）典型流程：

```go
// 1. 嵌入查询
embedResp, _ := embedModel.EmbedTexts(ctx, []string{"Go 并发最佳实践"})
queryVec := embedResp.Embeddings[0]

// 2. 向量检索
results, _ := store.SimilaritySearch(ctx, queryVec, 5,
    vectorstore.WithScoreThreshold(0.7),
    vectorstore.WithFilter(map[string]any{"lang": "zh"}),
)

// 3. 构建上下文
var context strings.Builder
for _, r := range results {
    context.WriteString(r.Document.Content + "\n")
}

// 4. 生成回答
resp, _ := chatModel.Generate(ctx, []ai.Message{
    ai.SystemMessage("根据以下文档回答问题：\n" + context.String()),
    ai.UserMessage("Go 并发最佳实践有哪些？"),
})
```

## 许可证

详见项目根目录 LICENSE 文件。

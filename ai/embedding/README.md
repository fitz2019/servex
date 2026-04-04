# ai/embedding

`ai/embedding` 包提供嵌入向量辅助工具函数：批量嵌入和余弦相似度计算。

## 功能特性

- `BatchEmbed`：将大批量文本按 batchSize 分批调用 `EmbeddingModel`，合并结果和 token 用量
- `CosineSimilarity`：计算两个向量的余弦相似度，返回值 [-1, 1]

## 安装

```bash
go get github.com/Tsukikage7/servex/ai
```

## API

```go
// 批量嵌入（自动分批）
func BatchEmbed(
    ctx context.Context,
    model ai.EmbeddingModel,
    texts []string,
    batchSize int,
    opts ...ai.CallOption,
) (*ai.EmbedResponse, error)

// 余弦相似度
func CosineSimilarity(a, b []float32) float32
```

## 使用示例

```go
client := openai.New(apiKey, openai.WithEmbeddingModel("text-embedding-3-small"))

// 批量嵌入 1000 条文档，每批 100 条
texts := loadDocuments() // 1000 条
resp, err := embedding.BatchEmbed(ctx, client, texts, 100)
if err != nil {
    return err
}
fmt.Printf("嵌入 %d 个向量，消耗 %d tokens\n",
    len(resp.Embeddings), resp.Usage.TotalTokens)

// 计算两个向量的相似度
query := resp.Embeddings[0]
for i, vec := range resp.Embeddings[1:] {
    score := embedding.CosineSimilarity(query, vec)
    fmt.Printf("文档 %d 相似度: %.4f\n", i+1, score)
}
```

## 许可证

详见项目根目录 LICENSE 文件。

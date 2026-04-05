# llm/retrieval/splitter

`github.com/Tsukikage7/servex/llm/retrieval/splitter` — 文本分块工具，支持按字符数、递归分隔符、Token 数三种分块策略，用于 RAG 文档预处理。

## 核心类型

- `Splitter` — 分块器接口，方法为 `Split(text) []Chunk`
- `Chunk` — 文本块，包含 Text、Offset（原文字节偏移）、Index（块序号）、Metadata
- `NewCharacterSplitter(opts...)` — 按字符数分块，相邻块保留重叠内容
- `NewRecursiveSplitter(opts...)` — 递归分块，优先按段落→句子→字符依次分割
- `NewTokenSplitter(opts...)` — 按估算 Token 数分块（ASCII 4 chars/token，CJK 1.5 chars/token）
- `WithChunkSize(n)` — 每块最大字符/Token 数（默认 1000）
- `WithChunkOverlap(n)` — 相邻块重叠字符/Token 数（默认 200）
- `WithSeparators(seps)` — 递归分块时使用的分隔符优先级列表

## 使用示例

```go
import "github.com/Tsukikage7/servex/llm/retrieval/splitter"

// 递归分块（推荐，自然分段）
sp := splitter.NewRecursiveSplitter(
    splitter.WithChunkSize(500),
    splitter.WithChunkOverlap(50),
)
chunks := sp.Split(longText)
for _, c := range chunks {
    fmt.Printf("[%d] offset=%d, len=%d\n", c.Index, c.Offset, len(c.Text))
}

// 按 Token 数分块
tokSp := splitter.NewTokenSplitter(
    splitter.WithChunkSize(256),
    splitter.WithChunkOverlap(32),
)
tokChunks := tokSp.Split(longText)
```

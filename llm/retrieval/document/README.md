# llm/retrieval/document

`github.com/Tsukikage7/servex/llm/retrieval/document` — 文档加载器，将各种文件格式转换为 `rag.Document`，供 RAG 管线使用。

## 核心类型

- `Loader` — 加载器接口，方法为 `Load(ctx) ([]rag.Document, error)`
- `NewTextLoader(reader, opts...)` — 从 `io.Reader` 加载单个文本文档
- `NewTextFileLoader(path, opts...)` — 从文件路径加载单个文本文档
- `NewCSVLoader(reader, opts...)` — 读取 CSV，每行生成一个文档（内容列默认 "content"）
- `NewJSONLoader(reader, opts...)` — 读取 JSON 数组或 JSONL，每个对象生成一个文档
- `NewMarkdownLoader(reader, opts...)` — 按 `##`/`###` 标题分节，每节生成一个文档
- `NewDirectoryLoader(dir, glob, opts...)` — 遍历目录，按 glob 模式匹配文件并加载
- `WithMetadata(meta)` — 附加自定义元数据
- `WithCSVContentColumn(col)` — 指定 CSV 内容列名
- `WithJSONContentField(field)` — 指定 JSON 内容字段名

## 使用示例

```go
import (
    "strings"
    "github.com/Tsukikage7/servex/llm/retrieval/document"
)

// 从字符串加载
loader := document.NewTextLoader(
    strings.NewReader("这是一段测试文本"),
    document.WithMetadata(map[string]any{"source": "manual"}),
)
docs, err := loader.Load(ctx)

// 从文件加载
fileLoader := document.NewTextFileLoader("/data/knowledge.txt")
docs, _ = fileLoader.Load(ctx)

// 目录加载
dirLoader := document.NewDirectoryLoader("/data/docs", "*.txt")
docs, _ = dirLoader.Load(ctx)

// Markdown 分节加载
mdLoader := document.NewMarkdownLoader(strings.NewReader(mdContent))
docs, _ = mdLoader.Load(ctx)
```

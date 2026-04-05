# llm/processing/structured

`github.com/Tsukikage7/servex/llm/processing/structured` — 结构化输出提取，将 LLM 约束为输出特定 Go struct，自动生成 JSON Schema 并支持失败重试。

## 核心类型

- `Extract[T any](ctx, model, prompt, opts...)` — 从单条 prompt 提取结构化数据，返回类型 T
- `ExtractFromMessages[T any](ctx, model, messages, opts...)` — 从已有消息列表提取结构化数据
- `SchemaFrom[T any]()` — 从 Go struct 类型参数生成 JSON Schema（支持 `json` tag 和 `description` tag）
- `WithMaxRetries(n)` — 设置 JSON 解析失败时的最大重试次数（默认 3）
- `WithCallOptions(opts...)` — 设置底层模型调用选项
- `WithSchemaDescription(desc)` — 设置 Schema 顶层描述

## 使用示例

```go
import "github.com/Tsukikage7/servex/llm/processing/structured"

type Product struct {
    Name     string  `json:"name" description:"产品名称"`
    Price    float64 `json:"price" description:"价格（元）"`
    Category string  `json:"category" description:"产品类别"`
}

product, err := structured.Extract[Product](
    ctx,
    myModel,
    "从以下文本中提取产品信息：iPhone 16 Pro，售价 8999 元，属于智能手机类别。",
)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("产品：%s，价格：%.2f，类别：%s\n",
    product.Name, product.Price, product.Category)
```

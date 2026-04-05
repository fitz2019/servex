# bizx/pagination — 游标分页

基于游标（Cursor-based）的分页实现，适合大数据量和实时数据（无跳页问题，性能稳定）。提供 GORM 集成辅助函数。

## 核心类型

```go
type CursorRequest struct {
    Cursor    string    // 上一页末尾游标，为空表示第一页
    Limit     int       // 每页数量，默认 20，最大 100
    Direction Direction // Forward（向后）或 Backward（向前），默认 Forward
}

type CursorResponse[T any] struct {
    Items      []T    // 本页数据
    NextCursor string // 下一页游标
    PrevCursor string // 上一页游标
    HasMore    bool   // 是否有更多数据
}
```

## 函数

| 函数 | 说明 |
|------|------|
| `EncodeCursor(values...)` | 将字段值编码为 base64url 游标字符串 |
| `DecodeCursor(cursor)` | 解码游标字符串，返回原始值列表 |
| `GORMPaginate(db, req, orderField)` | 为 GORM 查询自动添加游标条件和排序 |

## 常量

| 常量 | 值 | 说明 |
|------|----|------|
| `DefaultLimit` | 20 | 默认每页数量 |
| `MaxLimit` | 100 | 最大每页数量 |

## 快速上手

```go
// HTTP 处理器示例
func listPosts(w http.ResponseWriter, r *http.Request) {
    req := &pagination.CursorRequest{
        Cursor: r.URL.Query().Get("cursor"),
        Limit:  20,
    }.Apply() // Apply 归一化默认值

    var posts []Post
    db.Scopes(func(db *gorm.DB) *gorm.DB {
        return pagination.GORMPaginate(db, req, "id")
    }).Find(&posts)

    // 判断是否有更多
    hasMore := len(posts) > req.Limit
    if hasMore {
        posts = posts[:req.Limit]
    }

    // 生成下一页游标
    var nextCursor string
    if hasMore && len(posts) > 0 {
        nextCursor = pagination.EncodeCursor(posts[len(posts)-1].ID)
    }

    resp := &pagination.CursorResponse[Post]{
        Items:      posts,
        NextCursor: nextCursor,
        HasMore:    hasMore,
    }
    json.NewEncoder(w).Encode(resp)
}
```

## 错误

| 错误 | 说明 |
|------|------|
| `ErrInvalidCursor` | 游标格式无效（base64 解码失败或内容为空） |

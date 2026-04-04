---
name: storage
description: servex 存储模块专家。当用户使用 servex 的 storage/cache、storage/rdbms、storage/mongodb、storage/s3、storage/elasticsearch、storage/lock、storage/sqlx 时触发。
---

# servex 存储

## storage/cache — 缓存（Redis / 内存）

```go
// Redis 缓存
c, err := cache.NewCache(cache.NewRedisConfig("localhost:6379"), log)
if err != nil { ... }
defer c.Close()

// 写入（TTL 1 小时）
_ = c.Set(ctx, "user:1", `{"name":"alice"}`, time.Hour)

// 读取
val, err := c.Get(ctx, "user:1")

// 删除（方法名是 Del，不是 Delete）
_ = c.Del(ctx, "user:1")

// 内存缓存（开发/测试，无外部依赖）
memCache, err := cache.NewCache(cache.NewMemoryConfig(), log)
```

完整示例：`docs/superpowers/examples/storage/main.go`

**注意：** 删除方法是 `Del`（不是 `Delete`）。

## storage/rdbms — GORM 数据库

```go
// 支持 DriverPostgres / DriverMySQL / DriverSQLite / DriverClickHouse
db, err := database.NewDatabase(&database.Config{
    Driver:      database.DriverPostgres,
    DSN:         "host=localhost user=postgres password=postgres dbname=mydb port=5432 sslmode=disable",
    AutoMigrate: true,
}, log)
if err != nil { ... }
defer db.Close()

// 模型嵌入 BaseModel[T] 获得 ID、CreatedAt、UpdatedAt、DeletedAt（软删除）
// 字段名使用 `_at` 后缀：`CreatedAt`、`UpdatedAt`、`DeletedAt`
type User struct {
    database.BaseModel[uint]
    Name  string
    Email string
}

// 自动建表（仅 Config.AutoMigrate=true 时生效）
db.AutoMigrate(&User{})

// 获取带 context 的 *gorm.DB（推荐，确保链路追踪生效）
gdb := database.DB(ctx, db)

// 标准 GORM 操作
gdb.Create(&user)
gdb.First(&found, id)
gdb.Model(&found).Update("name", "new-name")
gdb.Delete(&found)  // 软删除
```

**注意：**
- `database.DB(ctx, db)` 推荐，而非 `db.AsGORM()`（后者无 context）
- `BaseModel` 字段名含 `Time` 后缀：`CreatedAt`、`UpdatedAt`、`DeletedAt`

## storage/mongodb — MongoDB 客户端

```go
// MustNewClient 初始化失败直接 panic（适合 main 函数）
client := mongodb.MustNewClient(mongodb.Config{
    URI:      "mongodb://localhost:27017",
    Database: "mydb",
})
defer client.Disconnect(ctx)

// NewClient 返回 error
client, err := mongodb.NewClient(mongodb.Config{...})
```

## storage/s3 — S3/MinIO 对象存储

```go
// MustNewClient 初始化失败直接 panic
s3 := s3.MustNewClient(s3.Config{
    Endpoint:  "localhost:9000",
    AccessKey: "minioadmin",
    SecretKey: "minioadmin",
    UseSSL:    false,
})

// NewClient 返回 error
s3, err := s3.NewClient(s3.Config{...})
```

## storage/elasticsearch — Elasticsearch 客户端

```go
// 创建 Elasticsearch 客户端
client, err := elasticsearch.NewClient(&elasticsearch.Config{
    Addresses: []string{"http://localhost:9200"},
    Username:  "elastic",
    Password:  "changeme",
    MaxRetries: 3,
    EnableTracing: true,
}, log)
if err != nil { ... }
defer client.Close()

// MustNewClient 失败时 panic
client := elasticsearch.MustNewClient(elasticsearch.DefaultConfig(), log)

// 索引操作
idx := client.Index("my-index")
idx.Create(ctx, map[string]any{
    "mappings": map[string]any{
        "properties": map[string]any{
            "name": map[string]any{"type": "text"},
        },
    },
})

// 文档 CRUD
doc := idx.Document()
doc.Index(ctx, "1", map[string]any{"name": "Alice"})
result, err := doc.Get(ctx, "1")
doc.Update(ctx, "1", map[string]any{"doc": map[string]any{"name": "Bob"}})
doc.Delete(ctx, "1")

// 批量操作
doc.Bulk(ctx, []elasticsearch.BulkAction{
    {Type: "index", ID: "1", Body: map[string]any{"name": "Alice"}},
    {Type: "delete", ID: "2"},
})

// 搜索
search := idx.Search()
results, err := search.Query(ctx, map[string]any{
    "match": map[string]any{"name": "Alice"},
}, elasticsearch.WithSize(10), elasticsearch.WithFrom(0))
```

**关键接口：**
- `elasticsearch.Client` — 顶层客户端（`Index`, `Ping`, `Close`, `Client`）
- `elasticsearch.Index` — 索引操作（`Create`, `Delete`, `Exists`, `Document`, `Search`）
- `elasticsearch.Document` — 文档 CRUD（`Index`, `Get`, `Update`, `Delete`, `Bulk`）
- `elasticsearch.Search` — 搜索（`Query`, `Count`, `Aggregate`, `Scroll`）
- 搜索选项：`WithSize`, `WithFrom`, `WithSort`, `WithHighlight`, `WithSourceIncludes`

## storage/lock — 分布式锁

```go
// 创建 Redis 分布式锁
locker := lock.NewRedis(cacheClient)

// 方式1: WithLock 辅助函数（推荐，自动获取/释放）
err := lock.WithLock(ctx, locker, "order:123", 30*time.Second, func() error {
    return processOrder(123)
})

// 方式2: TryWithLock 非阻塞版本
err := lock.TryWithLock(ctx, locker, "order:123", 30*time.Second, func() error {
    return processOrder(123)
})
if errors.Is(err, lock.ErrLockNotAcquired) {
    // 锁被占用
}

// 方式3: 手动管理
acquired, err := locker.TryLock(ctx, "my-resource", 30*time.Second)
if acquired {
    defer locker.Unlock(ctx, "my-resource")
    // ...
}

// 阻塞获取 + 延长锁
err := locker.Lock(ctx, "my-resource", 30*time.Second)
defer locker.Unlock(ctx, "my-resource")
locker.Extend(ctx, "my-resource", 30*time.Second) // 长操作时延长
```

**关键类型：**
- `lock.Locker` — 分布式锁接口（`TryLock`, `Lock`, `Unlock`, `Extend`）
- `lock.NewRedis(cacheClient)` — 基于 Redis 的实现
- `lock.WithLock(ctx, locker, key, ttl, fn)` — 自动管理锁生命周期
- `lock.TryWithLock(...)` — 非阻塞版本，获取失败立即返回
- 错误：`ErrLockNotAcquired`、`ErrLockNotHeld`、`ErrLockExpired`

## storage/sqlx — SQL 类型辅助工具

```go
// Nullable[T] 泛型 nullable 包装，支持 JSON/SQL 双向序列化
type User struct {
    Name  string
    Phone sqlx.Nullable[string] // 数据库中可为 NULL
    Age   sqlx.Nullable[int64]
}

// 创建 Nullable 值
phone := sqlx.Of("13800138000")  // Valid=true
noPhone := sqlx.Null[string]()   // Valid=false（NULL）

// 读取值（带默认值）
val := phone.ValueOr("未知")

// JSON 序列化：Valid=false → null，Valid=true → 值
// SQL 读写：自动实现 sql.Scanner 和 driver.Valuer

// 与标准库互转
ns := sqlx.NullableString(sql.NullString{String: "hello", Valid: true})
ni := sqlx.NullableInt64(sql.NullInt64{Int64: 42, Valid: true})
```

**关键类型：**
- `sqlx.Nullable[T]` — 泛型 nullable 包装（`Val T`, `Valid bool`）
- `sqlx.Of[T](v)` — 创建 Valid=true 的 Nullable
- `sqlx.Null[T]()` — 创建 Valid=false 的 Nullable（NULL）
- `n.ValueOr(def)` — 安全取值，NULL 时返回默认值
- 自动实现 `json.Marshaler`、`json.Unmarshaler`、`sql.Scanner`、`driver.Valuer`

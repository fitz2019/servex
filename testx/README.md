# testx

`github.com/Tsukikage7/servex/testx` — 测试辅助工具集，提供 Mock 日志、测试容器、HTTP/gRPC 测试服务器及 fixture 文件管理。

## 核心类型

- `NopLogger()` — 返回空操作日志记录器，实现 `logger.Logger` 但丢弃所有输出（适合单元测试）
- `TestLogger(t)` — 返回将日志输出到 `testing.T` 的日志记录器，测试失败时日志可见
- `Container` — 封装 testcontainers 容器，提供 `Addr()`、`Host()`、`Port()`、`Close()` 方法
- `NewRedis(ctx)` — 启动 Redis 测试容器（redis:7-alpine）
- `NewPostgres(ctx, opts...)` — 启动 PostgreSQL 测试容器（postgres:16-alpine），可配置用户/密码/库名
- `NewMySQL(ctx, opts...)` — 启动 MySQL 测试容器（mysql:8）
- `NewMongoDB(ctx)` — 启动 MongoDB 测试容器（mongo:7）
- `NewKafka(ctx)` — 启动 Kafka 测试容器（KRaft 模式）
- `NewClickHouse(ctx)` — 启动 ClickHouse 测试容器
- `HTTPTestServer` — 封装 `httptest.Server`，额外提供 `Get(path)`、`PostJSON(path, body)` 快捷方法
- `NewHTTPTestServer(handler, middlewares...)` — 创建 HTTP 测试服务器，支持中间件链
- `NewGRPCTestServer(registerFn, interceptors...)` — 创建基于内存连接（bufconn）的 gRPC 测试服务器，返回客户端连接和清理函数
- `LoadJSON[T](t, path)` — 从文件加载 JSON 并反序列化为指定类型
- `LoadYAML[T](t, path)` — 从文件加载 YAML 并反序列化为指定类型
- `Golden(t, name, actual)` — 对比 actual 与 golden 文件，`-update` 标志可更新 golden 文件
- `GoldenJSON(t, name, actual)` — 将 actual 序列化为格式化 JSON 后与 golden 文件对比

## 使用示例

```go
import "github.com/Tsukikage7/servex/testx"

func TestWithRedis(t *testing.T) {
    ctx := context.Background()

    // 启动 Redis 容器
    redis, err := testx.NewRedis(ctx)
    if err != nil {
        t.Fatal(err)
    }
    defer redis.Close(ctx)

    // 使用 redis.Addr() 连接
    client := redis.NewClient(&redis.Options{Addr: redis.Addr()})

    // 日志记录器
    log := testx.TestLogger(t)
    log.Infof("Redis 地址: %s", redis.Addr())
}

func TestHTTPHandler(t *testing.T) {
    srv := testx.NewHTTPTestServer(myHandler)
    defer srv.Close()

    resp := srv.PostJSON("/api/hello", map[string]string{"name": "world"})
    // 断言 resp ...
}

func TestGRPC(t *testing.T) {
    conn, cleanup := testx.NewGRPCTestServer(func(s *grpc.Server) {
        pb.RegisterMyServiceServer(s, &myService{})
    })
    defer cleanup()
    client := pb.NewMyServiceClient(conn)
    _ = client
}

// Golden 文件测试
func TestOutput(t *testing.T) {
    output := generateOutput()
    testx.GoldenJSON(t, "my_output", output)
}
```

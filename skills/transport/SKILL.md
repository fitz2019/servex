---
name: transport
description: servex 传输层模块专家。当用户使用 servex 的 httpserver、httpclient、grpcserver、ginserver、echoserver、hertzserver、websocket、sse、gateway、grpcclient、health、response、graphql 时触发，提供完整示例和用法。
---

# servex 传输层

## httpserver — 带认证的 HTTP 服务器

```go
// 适用场景：需要 JWT 认证、结构化日志、链路追踪的 HTTP 服务
srv := httpserver.New(mux,
    httpserver.WithLogger(log),
    httpserver.WithAddr(":8080"),
    httpserver.WithRecovery(),
    httpserver.WithLogging("/healthz"),           // 跳过健康检查路径的日志
    httpserver.WithTrace("my-service"),           // OpenTelemetry 服务名
    httpserver.WithAuth(authenticator, "/api/login"), // 公开路径白名单
)
if err := srv.Start(ctx); err != nil {
    log.Error(ctx, "启动失败", err)
}
```

完整示例：`docs/superpowers/examples/httpserver/main.go`

**关键选项：**
- `WithAddr(addr)` — 监听地址，默认 `:8080`
- `WithLogger(log)` — logger（必须）
- `WithRecovery()` — 捕获 panic，返回 500
- `WithLogging(skipPaths...)` — 结构化访问日志，跳过指定路径
- `WithTrace(serviceName)` — OpenTelemetry 中间件
- `WithAuth(authenticator, publicPaths...)` — 认证中间件，白名单外均需认证
- `WithMiddlewares(mws...)` — 注入自定义 `func(http.Handler) http.Handler`

## httpclient — 带负载均衡的 HTTP 客户端

```go
// 注意：WithServiceName、WithDiscovery、WithLogger 缺少任一将 panic（非 error 返回）
client, err := httpclient.New(
    httpclient.WithName("order-client"),        // 客户端标识，用于日志（可选）
    httpclient.WithLogger(log),
    httpclient.WithTimeout(5 * time.Second),
    httpclient.WithServiceName("order-service"), // 服务发现的 key（必须）
    httpclient.WithDiscovery(disc),              // discovery.Discovery 实现（consul/etcd/静态）
    httpclient.WithBalancer(&httpclient.RoundRobinBalancer{}),
)

// Do(ctx, method, path, body) — host 由 balancer/discovery 决定
resp, err := client.Do(ctx, http.MethodGet, "/api/orders", nil)
```

完整示例：`docs/superpowers/examples/httpclient/main.go`

**注意事项：**
- `httpclient.New` 构建时立即调用 `Discover`，地址列表不能为空
- 无真实注册中心时，可实现 `discovery.Discovery` 接口传入静态地址列表
- `RoundRobinBalancer` 是默认值，也可使用 `RandomBalancer`

## grpcserver — gRPC 服务器（片段）

```go
srv := grpcserver.New(
    grpcserver.WithLogger(log),
    grpcserver.WithAddr(":9090"),
    grpcserver.WithRecovery(),
    grpcserver.WithTrace("my-service"),
)
// 注册 gRPC 服务
pb.RegisterMyServiceServer(srv.Server(), &myServiceImpl{})
srv.Start(ctx)
```

## ginserver / echoserver / hertzserver — 框架适配（片段）

```go
// Gin
engine := gin.New()
ginserver.New(engine).ApplyMiddlewares(
    ginserver.Recovery(),
    ginserver.RequestID(),
)

// Echo
e := echo.New()
echoserver.New(e).ApplyMiddlewares(
    echoserver.Recovery(),
)
```

## websocket — WebSocket 服务端（片段）

```go
handler := websocket.NewHandler(
    websocket.WithOnConnect(func(conn *websocket.Conn) {
        log.Info("新连接:", conn.ID())
    }),
    websocket.WithOnMessage(func(conn *websocket.Conn, msg []byte) {
        conn.Send(msg) // echo
    }),
    websocket.WithOnDisconnect(func(conn *websocket.Conn, err error) {}),
)

mux.Handle("/ws", handler)
```

## sse — Server-Sent Events（片段）

```go
handler := sse.NewHandler(
    sse.WithOnConnect(func(client *sse.Client) {
        client.Send(&sse.Event{Data: "connected"})
    }),
)

mux.Handle("/events", handler)
// 向所有客户端广播
handler.Broadcast(&sse.Event{Event: "update", Data: payload})
```

## gateway — gRPC + HTTP 双协议服务器（gRPC-Gateway）

```go
// 创建 Gateway 服务器（同时暴露 gRPC 和 HTTP 端口）
srv := gateway.New(
    gateway.WithName("order-service"),
    gateway.WithGRPCAddr(":9090"),
    gateway.WithHTTPAddr(":8080"),
    gateway.WithLogger(log),
    gateway.WithRecovery(),                    // panic 恢复（双端）
    gateway.WithTrace("order-service"),         // 链路追踪（双端）
    gateway.WithResponse(),                     // 统一响应格式
    gateway.WithReflection(true),               // gRPC 反射
    gateway.WithAuth(authenticator),            // 认证（双端）
    gateway.WithPublicMethods(
        "/api.order.v1.OrderService/List",      // 精确匹配
        "/api.auth.v1.AuthService/*",           // 服务级通配
    ),
    gateway.WithAutoDiscovery(),                // 从 proto option 自动发现公开方法
    gateway.WithReadinessChecker(dbChecker),    // 就绪检查
)

// 注册服务（需实现 gateway.Registrar 接口）
srv.Register(&OrderService{}, &UserService{})

// 启动
if err := srv.Start(ctx); err != nil { ... }
defer srv.Stop(ctx)
```

**关键类型：**
- `gateway.New(opts...) *Server` — 构造器
- `gateway.Registrar` — 服务注册接口（`RegisterGRPC` + `RegisterGateway`）
- `server.Register(services...)` — 注册业务服务
- 内置健康检查：`/healthz`（存活）、`/readyz`（就绪）
- `WithConfig(transport.GatewayConfig)` — 从配置结构体设置

## grpcclient — gRPC 客户端

```go
// 创建 gRPC 客户端（通过服务发现获取地址）
client, err := grpcclient.New(
    grpcclient.WithName("order-client"),
    grpcclient.WithServiceName("order-service"),  // 必需
    grpcclient.WithDiscovery(disc),               // 必需
    grpcclient.WithLogger(log),                   // 必需
    grpcclient.WithInterceptors(tracingInterceptor),
    grpcclient.WithDialOptions(grpc.WithBlock()),
)
if err != nil { ... }
defer client.Close()

// 获取底层 gRPC 连接，创建 stub
conn := client.Conn()
orderClient := pb.NewOrderServiceClient(conn)
resp, err := orderClient.GetOrder(ctx, &pb.GetOrderRequest{Id: "1"})
```

**关键类型：**
- `grpcclient.New(opts...) (*Client, error)` — 构造器（serviceName、discovery、logger 必需，缺少会 panic）
- `client.Conn() *grpc.ClientConn` — 获取底层连接
- `WithInterceptors(...)` — 添加一元拦截器
- `WithDialOptions(...)` — 额外 dial 选项

## health — 健康检查

```go
// 创建健康检查管理器
h := health.New(
    health.WithTimeout(5 * time.Second),
    health.WithLivenessChecker(health.NewAlwaysUpChecker("self")),
    health.WithReadinessChecker(
        health.NewDBChecker("postgres", dbPinger),
        health.NewRedisChecker("redis", redisPinger),
    ),
)

// 动态添加检查器
h.AddReadinessChecker(health.NewPingChecker("es", esPinger))

// 自定义检查器
h.AddReadinessChecker(health.NewCheckerFunc("custom", func(ctx context.Context) health.CheckResult {
    if err := doCheck(); err != nil {
        return health.CheckResult{Status: health.StatusDown, Message: err.Error()}
    }
    return health.CheckResult{Status: health.StatusUp}
}))

// 注册 HTTP 路由（/healthz + /readyz）
handler := health.NewHTTPHandler(h)
handler.RegisterRoutes(mux)

// 或使用中间件（自动拦截 /healthz、/readyz）
srv := httpserver.New(health.Middleware(h)(mux))

// 判断是否健康
if h.IsHealthy(ctx) { ... }
```

**关键类型：**
- `health.Health` — 健康检查管理器（`Liveness`, `Readiness`, `IsHealthy`）
- `health.Checker` — 检查器接口（`Name() string`, `Check(ctx) CheckResult`）
- 内置检查器：`NewDBChecker`、`NewRedisChecker`、`NewPingChecker`、`NewAlwaysUpChecker`、`NewCompositeChecker`
- `health.Middleware(h)` — HTTP 中间件，自动拦截 `/healthz`、`/readyz`
- 状态：`StatusUp`、`StatusDown`、`StatusUnknown`

## response — 统一响应格式

```go
// 统一响应体：{"code": 0, "message": "成功", "data": {...}}
resp := response.OK(user)                           // 成功
resp := response.OKWithMessage(user, "创建成功")
resp := response.Fail[any](response.CodeNotFound)    // 失败
resp := response.FailWithMessage[any](response.CodeInvalidParam, "ID 不能为空")
resp := response.FailWithError[any](err)             // 从 error 提取错误码

// 分页响应：{"code": 0, "message": "成功", "data": [...], "pagination": {...}}
resp := response.Paged(paginationResult)

// 业务错误
err := response.NewError(response.CodeNotFound)
err := response.NewErrorWithMessage(response.CodeInvalidParam, "用户名已存在")
err := response.Wrap(response.CodeDatabaseError, dbErr)

// 提取错误信息
code := response.ExtractCode(err)      // 提取错误码
msg := response.ExtractMessage(err)     // 提取消息（5xxxx 隐藏详情）
msg := response.ExtractMessageUnsafe(err) // 完整消息（仅用于日志）
```

**错误码规范：**
- `0` — 成功
- `1xxxx` — 通用错误（`CodeUnknown`, `CodeCanceled`, `CodeTimeout`）
- `2xxxx` — 认证/授权（`CodeUnauthorized`, `CodeForbidden`, `CodeTokenExpired`）
- `3xxxx` — 参数（`CodeInvalidParam`, `CodeMissingParam`, `CodeValidationFailed`）
- `4xxxx` — 资源（`CodeNotFound`, `CodeAlreadyExists`, `CodeConflict`）
- `5xxxx` — 内部（`CodeInternal`, `CodeDatabaseError`）
- `6xxxx` — 外部服务（`CodeServiceUnavailable`, `CodeUpstreamError`）

**关键类型：**
- `response.Response[T]` / `response.PagedResponse[T]` — 统一响应体（实现 `Envelope` 接口）
- `response.Code` — 错误码（含 `Num`、`Message`、`HTTPStatus`、`GRPCCode`、`Key`）
- `response.BusinessError` — 业务错误（含 `Code`、`Message`、`Cause`）
- `response.NewCode(num, message, httpStatus, grpcCode)` — 自定义错误码

## transport/graphql — GraphQL 服务器适配

```go
// 1. 定义 GraphQL Schema（使用 graphql-go/graphql）
userType := graphql.NewObject(graphql.ObjectConfig{
    Name: "User",
    Fields: graphql.Fields{
        "id":   &graphql.Field{Type: graphql.String},
        "name": &graphql.Field{Type: graphql.String},
    },
})

queryType := graphql.NewObject(graphql.ObjectConfig{
    Name: "Query",
    Fields: graphql.Fields{
        "user": &graphql.Field{
            Type: userType,
            Args: graphql.FieldConfigArgument{
                "id": &graphql.ArgumentConfig{Type: graphql.String},
            },
            // 使用 WrapResolve 为单个字段添加中间件
            Resolve: gqlserver.WrapResolve(
                func(p graphql.ResolveParams) (any, error) {
                    id, _ := p.Args["id"].(string)
                    return findUser(p.Context, id)
                },
                gqlserver.LoggingMiddleware(log),
                gqlserver.TracingMiddleware("user-service"),
            ),
        },
    },
})

schema, _ := graphql.NewSchema(graphql.SchemaConfig{Query: queryType})

// 2. 创建 GraphQL 服务器
srv := gqlserver.New(schema,
    gqlserver.WithLogger(log),
    gqlserver.WithConfig(&gqlserver.Config{
        Pretty:     false,
        Playground: true,     // 启用 GraphiQL UI
        Path:       "/graphql",
    }),
    // 全局中间件（对所有 resolve 函数生效）
    gqlserver.WithMiddleware(
        gqlserver.RecoveryMiddleware(log),
        gqlserver.LoggingMiddleware(log),
        gqlserver.TracingMiddleware("my-service"),
    ),
)

// 3. 注册路由
mux.Handle("/graphql", srv.Handler())
if cfg.Playground {
    mux.Handle("/playground", srv.PlaygroundHandler())
}
```

**内置中间件（resolve 层）：**

```go
// 日志：记录字段名和耗时
gqlserver.LoggingMiddleware(log)

// 链路追踪：为每次 resolve 创建 OTel span
gqlserver.TracingMiddleware("service-name")

// Panic 恢复：防止单个 resolve 崩溃整个服务
gqlserver.RecoveryMiddleware(log)

// 链接多个中间件
combined := gqlserver.ChainMiddleware(
    gqlserver.RecoveryMiddleware(log),
    gqlserver.LoggingMiddleware(log),
    gqlserver.TracingMiddleware("svc"),
)
```

**关键类型：**
- `graphql.New(schema, opts...) *Server` — 创建服务器
- `server.Handler() http.Handler` — GraphQL 请求处理器（支持 GET/POST）
- `server.PlaygroundHandler() http.Handler` — GraphiQL 交互式 UI
- `graphql.Config` — 配置（`Pretty bool`, `Playground bool`, `Path string`）
- `graphql.Middleware` — `func(ResolveFunc) ResolveFunc`，resolve 层中间件
- `graphql.WrapResolve(fn, mw...)` — 将中间件应用到单个 resolve 函数
- `graphql.ChainMiddleware(outer, others...)` — 链接多个中间件
- 内置中间件：`LoggingMiddleware(log)`, `TracingMiddleware(serviceName)`, `RecoveryMiddleware(log)`
- `graphql.ErrorHandlerFunc` — `func(ctx, []gqlerrors.FormattedError) []gqlerrors.FormattedError`，自定义错误处理

**注意：**
- Schema 定义使用 `github.com/graphql-go/graphql`，servex 提供服务器适配层
- `WithMiddleware` 为全局中间件（所有 resolve），`WrapResolve` 为字段级中间件
- `DefaultConfig()` 默认启用 Playground，路径为 `/graphql`

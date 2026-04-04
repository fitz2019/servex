# servex

Go 微服务开发工具包，提供构建生产级微服务所需的核心组件。

## 安装

```bash
go get github.com/Tsukikage7/servex
```

## Claude Code Plugin

servex 内置 [Claude Code Plugin](https://code.claude.com/docs/en/plugins.md)，为 AI 辅助开发提供 21 个专业 skill（模块使用指南、代码生成规范、最佳实践）。

**安装插件：**

```bash
claude plugin add --from https://github.com/Tsukikage7/servex
```

**可用 Skills（21 个）：**

| Skill | 说明 |
|-------|------|
| `servex:servex` | 主索引 — 模块映射、代码生成规范、工作流程 |
| `servex:transport` | 传输层（httpserver/grpcserver/httpclient/gateway/WS/SSE） |
| `servex:middleware` | 中间件（限流/熔断/重试/恢复/超时/CORS/幂等/并发控制） |
| `servex:auth` | 认证（JWT/API Key） |
| `servex:storage` | 存储（Cache/RDBMS/MongoDB/Elasticsearch/S3/Lock/SQLx） |
| `servex:observability` | 可观测性（Metrics/Tracing/Logger） |
| `servex:config` | 配置与服务发现（File/Etcd/Consul/Env） |
| `servex:ai` | AI（OpenAI/Anthropic/Gemini/ToolCall/Router/Embedding/Prompt） |
| `servex:distributed` | 分布式模式（CQRS/Outbox/Saga/领域事件） |
| `servex:pubsub` | 消息与任务（Pub/Sub/JobQueue） |
| `servex:webhook` | Webhook 投递与接收 |
| `servex:oauth2` | OAuth2 第三方登录（GitHub/Google/微信） |
| `servex:openapi` | OpenAPI 文档生成 |
| `servex:collections` | 数据结构（Deque/LRU/TreeMap/PriorityQueue/HashSet 等） |
| `servex:xutil` | 工具包（ptrx/strx/copier/syncx/sorting/pagination 等） |
| `servex:httpx` | HTTP 请求分析（ClientIP/UserAgent/Locale/BotDetect 等） |
| `servex:errors` | 统一错误处理（HTTP/gRPC 映射） |
| `servex:i18n` | 国际化 |
| `servex:tenant` | 多租户（中间件 + GORM Scope） |
| `servex:notify` | 通知系统（Email/SMS/Push/Webhook 渠道） |
| `servex:testx` | 测试工具包（NopLogger/TestLogger/Container/HTTPTest/Fixture） |

安装后 Claude 会根据上下文自动触发对应 skill，也可手动调用如 `/servex:storage`。

## 包概览

### 核心

| 包 | 说明 |
| --- | --- |
| [app](./app/) | 应用生命周期管理 |
| [endpoint](./endpoint/) | Endpoint / Middleware 核心抽象 |
| [errors](./errors/) | 统一错误（HTTP/gRPC 状态码映射） |
| [encoding](./encoding/) | 编解码器接口与 HTTP 内容协商（json/proto/xml/pbjson） |

### 传输层 (transport/)

| 包 | 说明 |
| --- | --- |
| [transport/httpserver](./transport/httpserver/) | HTTP 服务器（pprof、Recovery、中间件） |
| [transport/grpcserver](./transport/grpcserver/) | gRPC 服务器 |
| [transport/httpclient](./transport/httpclient/) | HTTP 客户端（Config 驱动、retry/circuitbreaker/tracing/metrics 内置中间件） |
| [transport/grpcclient](./transport/grpcclient/) | gRPC 客户端 |
| [transport/gateway](./transport/gateway/) | API 网关 |
| [transport/ginserver](./transport/ginserver/) | Gin 适配 |
| [transport/echoserver](./transport/echoserver/) | Echo 适配 |
| [transport/hertzserver](./transport/hertzserver/) | Hertz 适配 |
| [transport/websocket](./transport/websocket/) | WebSocket 服务端（gorilla/websocket） |
| [transport/sse](./transport/sse/) | Server-Sent Events 服务端 |
| [transport/health](./transport/health/) | 健康检查（K8s 探针） |
| [transport/response](./transport/response/) | 统一响应封装 |
| [transport/graphql](./transport/graphql/) | GraphQL 服务器适配（graphql-go/graphql） |

### 中间件 (middleware/)

| 包 | 说明 | Endpoint | HTTP | gRPC |
| --- | --- | :---: | :---: | :---: |
| [middleware/ratelimit](./middleware/ratelimit/) | 限流（令牌桶、滑动窗口、分布式） | Y | Y | Y |
| [middleware/circuitbreaker](./middleware/circuitbreaker/) | 熔断器（Closed/Open/HalfOpen） | Y | Y | Y |
| [middleware/retry](./middleware/retry/) | 重试机制（指数退避） | Y | Y | Y |
| [middleware/recovery](./middleware/recovery/) | Panic 恢复 | Y | Y | Y |
| [middleware/timeout](./middleware/timeout/) | 超时控制 | Y | Y | Y |
| [middleware/cors](./middleware/cors/) | 跨域资源共享（CORS） | - | Y | - |
| [middleware/requestid](./middleware/requestid/) | 请求 ID 注入与传播 | Y | Y | Y |
| [middleware/idempotency](./middleware/idempotency/) | 幂等性保证 | Y | Y | - |
| [middleware/semaphore](./middleware/semaphore/) | 并发控制 | Y | - | - |
| [middleware/logging](./middleware/logging/) | 请求日志（HTTP / gRPC） | - | Y | Y |

### 认证 (auth/)

| 包 | 说明 |
| --- | --- |
| [auth/jwt](./auth/jwt/) | JWT 认证（签发/验证/白名单） |
| [auth/apikey](./auth/apikey/) | API Key 认证 |

### 可观测性 (observability/)

| 包 | 说明 | Endpoint | HTTP | gRPC |
| --- | --- | :---: | :---: | :---: |
| [observability/logger](./observability/logger/) | 结构化日志（Zap） | - | - | - |
| [observability/metrics](./observability/metrics/) | Prometheus 指标收集 | Y | Y | Y |
| [observability/tracing](./observability/tracing/) | OpenTelemetry 链路追踪 | Y | Y | Y |

### 配置与服务发现

| 包 | 说明 |
| --- | --- |
| [config](./config/) | 配置管理（多源热加载、Source 抽象） |
| [config/source/file](./config/source/file/) | 文件配置源 |
| [config/source/etcd](./config/source/etcd/) | etcd 配置源 |
| [config/source/consul](./config/source/consul/) | Consul 配置源 |
| [config/source/env](./config/source/env/) | 环境变量配置源 |
| [discovery](./discovery/) | 服务发现（Consul、etcd） |

### 存储 (storage/)

| 包 | 说明 | 工厂函数 |
| --- | --- | --- |
| [storage/cache](./storage/cache/) | 缓存（内存、Redis） | `NewCache` / `MustNewCache` |
| [storage/rdbms](./storage/rdbms/) | 关系数据库（GORM） | `NewDatabase` / `MustNewDatabase` |
| [storage/mongodb](./storage/mongodb/) | MongoDB 客户端 | `NewClient` / `MustNewClient` |
| [storage/elasticsearch](./storage/elasticsearch/) | Elasticsearch 客户端 | `NewClient` / `MustNewClient` |
| [storage/s3](./storage/s3/) | S3 兼容对象存储 | `NewClient` / `MustNewClient` |
| [storage/lock](./storage/lock/) | 分布式锁 | `NewLock` |
| [storage/sqlx](./storage/sqlx/) | sqlx 封装 | `NewDB` |
| [storage/migration](./storage/migration/) | 数据库迁移（Go DSL） | `NewRegistry` / `NewRunner` |
| [storage/clickhouse](./storage/clickhouse/) | ClickHouse 客户端 | `NewClient` / `MustNewClient` |

### 消息 (messaging/)

| 包 | 说明 | 工厂函数 |
| --- | --- | --- |
| [messaging/pubsub](./messaging/pubsub/) | 统一 Pub/Sub 抽象 | - |
| [messaging/pubsub/factory](./messaging/pubsub/factory/) | **Config 驱动工厂（推荐）** | `NewPublisher` / `NewSubscriber` |
| [messaging/pubsub/kafka](./messaging/pubsub/kafka/) | Kafka driver | `NewPublisher` / `NewSubscriber` |
| [messaging/pubsub/rabbitmq](./messaging/pubsub/rabbitmq/) | RabbitMQ driver | `NewPublisher` / `NewSubscriber` |
| [messaging/pubsub/redis](./messaging/pubsub/redis/) | Redis Streams driver | `NewPublisher` / `NewSubscriber` |
| [messaging/jobqueue](./messaging/jobqueue/) | 异步任务队列（延迟、优先级、重试、死信） | `NewClient` / `NewWorker` |
| [messaging/jobqueue/factory](./messaging/jobqueue/factory/) | **Config 驱动工厂（推荐）** | `NewStore` |
| [messaging/jobqueue/redis](./messaging/jobqueue/redis/) | Redis Store | `NewStore` |
| [messaging/jobqueue/kafka](./messaging/jobqueue/kafka/) | Kafka Store | `NewStore` |
| [messaging/jobqueue/rabbitmq](./messaging/jobqueue/rabbitmq/) | RabbitMQ Store | `NewStore` |

### 领域驱动 (domain/)

| 包 | 说明 |
| --- | --- |
| [domain](./domain/) | 聚合根、领域事件、EventBus |
| [domain/cqrs](./domain/cqrs/) | 命令查询职责分离 |
| [domain/saga](./domain/saga/) | Saga 分布式事务 |
| [domain/outbox](./domain/outbox/) | 事务发件箱模式 |
| [domain/eventsourcing](./domain/eventsourcing/) | 事件溯源（Event Sourcing） |

### 通知 (notify/)

| 包 | 说明 |
| --- | --- |
| [notify](./notify/) | 统一通知接口 |
| [notify/email](./notify/email/) | 邮件通知 |
| [notify/sms](./notify/sms/) | 短信通知 |
| [notify/push](./notify/push/) | 推送通知 |
| [notify/webhook](./notify/webhook/) | Webhook 投递与接收 |
| [notify/nwebhook](./notify/nwebhook/) | Webhook 通知渠道 |
| [notify/factory](./notify/factory/) | 通知渠道工厂 |

### HTTP 请求分析 (httpx/)

| 包 | 说明 |
| --- | --- |
| [httpx](./httpx/) | 组合中间件（统一入口） |
| [httpx/clientip](./httpx/clientip/) | 客户端 IP 提取、地理位置、ACL |
| [httpx/useragent](./httpx/useragent/) | User-Agent 解析 |
| [httpx/deviceinfo](./httpx/deviceinfo/) | 设备信息（Client Hints 优先） |
| [httpx/botdetect](./httpx/botdetect/) | 机器人检测 |
| [httpx/locale](./httpx/locale/) | 语言区域设置 |
| [httpx/referer](./httpx/referer/) | 来源页面解析、UTM 参数 |
| [httpx/activity](./httpx/activity/) | 用户活动追踪（Redis + Kafka） |

### AI 集成 (ai/)

| 包 | 说明 |
| --- | --- |
| [ai](./ai/) | 统一 ChatModel / EmbeddingModel 接口抽象 |
| [ai/openai](./ai/openai/) | OpenAI 适配器（兼容 DeepSeek、通义千问等） |
| [ai/anthropic](./ai/anthropic/) | Anthropic Claude 适配器 |
| [ai/gemini](./ai/gemini/) | Google Gemini 适配器 |
| [ai/router](./ai/router/) | 多 Provider 路由器（按模型名路由） |
| [ai/toolcall](./ai/toolcall/) | 工具注册与自动 ReAct 循环执行器 |
| [ai/conversation](./ai/conversation/) | 多轮对话会话管理（BufferMemory / WindowMemory） |
| [ai/embedding](./ai/embedding/) | 批量嵌入 + 余弦相似度工具函数 |
| [ai/prompt](./ai/prompt/) | 基于 text/template 的提示词模板引擎 |
| [ai/middleware](./ai/middleware/) | AI 中间件链（日志、重试、限流、用量追踪） |
| [ai/vectorstore](./ai/vectorstore/) | 向量存储统一接口抽象 |

### OAuth2 第三方登录

| 包 | 说明 |
| --- | --- |
| [oauth2](./oauth2/) | Provider / StateStore 接口 |
| [oauth2/github](./oauth2/github/) | GitHub OAuth2 Provider |
| [oauth2/google](./oauth2/google/) | Google OAuth2 Provider |
| [oauth2/wechat](./oauth2/wechat/) | 微信 OAuth2 Provider |
| [oauth2/state](./oauth2/state/) | State 管理（Memory / Redis） |

### 其他

| 包 | 说明 |
| --- | --- |
| [openapi](./openapi/) | Code-first OpenAPI 3.0 生成 |
| [scheduler](./scheduler/) | Cron 定时任务调度 |
| [i18n](./i18n/) | 国际化 |
| [tenant](./tenant/) | 多租户（GORM Scope） |
| [collections](./collections/) | 数据结构（Deque/LRU/TreeMap/PriorityQueue/HashSet 等，12 子包） |
| [xutil](./xutil/) | 工具包（ptrx/strx/randx/iox/copier/syncx/sorting/pagination/version/crypto/optionx/valuex） |
| [testx](./testx/) | 测试工具包（NopLogger/TestLogger/Container/HTTPTest/Fixture） |

## 设计原则

- **KISS** - 保持简单，避免过度设计
- **DRY** - 抽象通用模式，减少重复代码
- **SOLID** - 单一职责，接口隔离
- **可组合** - 中间件可自由组合，支持 Endpoint / HTTP / gRPC 三层
- **可扩展** - 所有核心组件基于接口，支持自定义实现

## 工厂函数命名规范

| 模式 | 说明 | 示例 |
| --- | --- | --- |
| `NewXXX` | 返回 `(T, error)` | `rdbms.NewDatabase(cfg, log)` |
| `MustNewXXX` | 失败时 panic | `rdbms.MustNewDatabase(cfg, log)` |
| `DefaultConfig` | 返回默认配置 | `logger.DefaultConfig()` |

## 中间件执行顺序

推荐顺序（从外到内）：RequestID → Logging → Tracing → Metrics → RateLimit → CircuitBreaker → Retry → Timeout → Recovery

详见各中间件包的 README。

## License

MIT License

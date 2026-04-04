---
name: servex
description: servex Go 微服务工具库专家。当用户在使用 servex（项目依赖 github.com/Tsukikage7/servex）时触发，提供模块索引、代码生成规范和工作流程。
---

# servex 使用指南

## 触发条件

**触发：**
- 用户描述业务意图，如"我需要一个带熔断的 HTTP 服务"
- 用户明确说出 servex 模块名，如"帮我用 servex 的 circuitbreaker"
- 用户在 servex 项目中问某功能怎么实现

**不触发：**
- 问题涉及其他库（go-zero、kratos、gin 等），servex 只是对比提及
- 纯粹的 Go 语言问题，与 servex API 无关
- 无法确认用户项目使用 servex 时，先询问确认，而非直接生成

## 工作流程

1. **业务意图（模糊描述）** → 映射到合适的 servex 模块（1-2 句说明理由）→ 生成完整可运行示例
2. **模块名明确** → 直接生成该模块的示例或片段
3. **需求模糊，无法映射** → 先问一个澄清问题，再生成
4. **不确定 API 细节**（函数签名、选项名、默认值）→ 读取 servex 源码对应文件，不猜测
5. **生成多中间件代码** → 严格按此顺序：requestid → logging → tracing → metrics → ratelimit → circuitbreaker → retry → timeout → recovery

**源码定位规则：**
- 若在 servex 仓库本身，源码根目录即为当前目录
- 否则在 go.mod 的 `replace` 指令或 module cache 中找到源码
- 按包路径直接定位目录，如不确定 `middleware/circuitbreaker/` 的选项，读取该目录下的 `.go` 文件

## 代码生成规范

- **完整示例**（`main.go` 级别）：用户需求模糊/业务层面时生成
- **片段**（函数/结构体级别）：用户需求具体/技术层面时生成
- 注释中文，选项模式（`WithXxx`），错误处理显式，不 panic
- 输出格式：先说明选择的模块及原因（1-2 句），再给出代码，最后附关键配置项说明

## 模块索引

### 传输层 → 详见 `transport` skill

| 模块 | 包路径 | 描述 | 核心类型/函数 |
|------|--------|------|--------------|
| httpserver | `transport/httpserver` | HTTP 服务器（集成中间件链） | `New`, `WithAddr`, `WithLogger`, `WithAuth`, `WithMiddlewares` |
| grpcserver | `transport/grpcserver` | gRPC 服务器 | `New`, `Server` |
| httpclient | `transport/httpclient` | HTTP 客户端（负载均衡） | `New`, `WithServiceName`, `WithDiscovery`, `WithBalancer` |
| ginserver | `transport/ginserver` | Gin 适配器 | `New` |
| echoserver | `transport/echoserver` | Echo 适配器 | `New` |
| hertzserver | `transport/hertzserver` | Hertz 适配器 | `New` |
| websocket | `transport/websocket` | WebSocket 服务端 | `NewServer`, `Handler` |
| sse | `transport/sse` | Server-Sent Events 服务端 | `NewServer`, `Handler` |
| gateway | `transport/gateway` | gRPC + HTTP 双协议服务器 | `New`, `Register`, `Registrar`, `WithAuth`, `WithPublicMethods` |
| grpcclient | `transport/grpcclient` | gRPC 客户端 | `New`, `Conn`, `WithServiceName`, `WithDiscovery` |
| health | `transport/health` | 健康检查 | `New`, `Checker`, `NewDBChecker`, `NewRedisChecker`, `Middleware` |
| response | `transport/response` | 统一响应格式 | `OK`, `Fail`, `Response`, `Code`, `BusinessError`, `ExtractCode` |

### 中间件 → 详见 `middleware` skill

| 模块 | 包路径 | 描述 | 核心类型/函数 |
|------|--------|------|--------------|
| ratelimit | `middleware/ratelimit` | 限流（令牌桶、滑动窗口） | `NewTokenBucket`, `NewSlidingWindow`, `HTTPMiddleware` |
| circuitbreaker | `middleware/circuitbreaker` | 熔断器（Closed/Open/HalfOpen） | `New`, `WithFailureThreshold`, `WithOpenTimeout`, `HTTPMiddleware` |
| retry | `middleware/retry` | 重试（指数退避） | `New`, `WithMaxAttempts`, `WithBackoff` |
| recovery | `middleware/recovery` | Panic 恢复 | `New` |
| timeout | `middleware/timeout` | 超时控制 | `New`, `WithTimeout` |
| cors | `middleware/cors` | 跨域 | `New`, `WithAllowOrigins` |
| requestid | `middleware/requestid` | 请求 ID 注入 | `New`, `FromContext` |
| idempotency | `middleware/idempotency` | 幂等性保证 | `New`, `WithStore` |
| semaphore | `middleware/semaphore` | 并发控制 | `New`, `WithLimit` |
| logging | `middleware/logging` | 结构化请求日志 | `NewHTTP`, `NewGRPC` |

### 认证 → 详见 `auth` skill

| 模块 | 包路径 | 描述 | 核心类型/函数 |
|------|--------|------|--------------|
| auth/jwt | `auth/jwt` | JWT 签发与验证 | `NewJWT`, `NewAuthenticator`, `WithSecretKey`, `Generate`, `Validate` |
| auth/apikey | `auth/apikey` | API Key 验证 | `New`, `StaticValidator`, `CacheValidator` |

### 存储 → 详见 `storage` skill

| 模块 | 包路径 | 描述 | 核心类型/函数 |
|------|--------|------|--------------|
| storage/cache | `storage/cache` | 缓存（内存、Redis） | `NewCache`, `NewRedisConfig`, `NewMemoryConfig` |
| storage/rdbms | `storage/rdbms` | 数据库（GORM） | `NewDatabase`, `BaseModel`, `DB`, `AsGORM` |
| storage/mongodb | `storage/mongodb` | MongoDB 客户端 | `NewClient`, `MustNewClient` |
| storage/s3 | `storage/s3` | S3/MinIO 对象存储 | `NewClient`, `MustNewClient` |
| storage/elasticsearch | `storage/elasticsearch` | Elasticsearch 客户端 | `NewClient`, `MustNewClient` |
| storage/lock | `storage/lock` | 分布式锁 | `NewLocker`, `Lock`, `Unlock` |
| storage/sqlx | `storage/sqlx` | sqlx 数据库封装 | `NewDB`, `MustNewDB` |

### 可观测性 → 详见 `observability` skill

| 模块 | 包路径 | 描述 | 核心类型/函数 |
|------|--------|------|--------------|
| observability/metrics | `observability/metrics` | Prometheus 指标 | `NewMetrics`, `MustNewMetrics`, `DefaultConfig` |
| observability/tracing | `observability/tracing` | OpenTelemetry 追踪 | `NewTracer`, `TracingConfig`, `OTLPConfig` |
| observability/logger | `observability/logger` | 结构化日志 | `NewLogger`, `WithLevel`, `WithOutput` |

### 配置与服务发现 → 详见 `config` skill

| 模块 | 包路径 | 描述 | 核心类型/函数 |
|------|--------|------|--------------|
| config | `config` | 多源配置管理 | `NewManager`, `WithSource`, `Load`, `Watch`, `Get` |
| config/source/file | `config/source/file` | 文件配置源（热更新） | `New` |
| config/source/etcd | `config/source/etcd` | etcd 配置源 | `New` |
| config/source/env | `config/source/env` | 环境变量配置源 | `New` |
| config/source/consul | `config/source/consul` | Consul KV 配置源 | `New`, `WithFormat`, `WithDatacenter` |
| discovery | `discovery` | 服务注册与发现 | `NewDiscovery`, `NewServiceRegistry`, `Register`, `Discover` |

### AI → 详见 `ai` skill

| 模块 | 包路径 | 描述 | 核心类型/函数 |
|------|--------|------|--------------|
| ai/toolcall | `ai/toolcall` | AI 工具调用框架 | `NewRegistry`, `NewExecutor`, `WithOnStep` |
| ai/router | `ai/router` | 多 Provider 路由 | `New`, `Route` |
| ai/openai | `ai/openai` | OpenAI 客户端（兼容 DeepSeek 等） | `New`, `WithBaseURL`, `WithModel`, `WithEmbeddingModel` |
| ai/anthropic | `ai/anthropic` | Anthropic Claude 客户端 | `New`, `WithModel`, `WithDefaultMaxTokens` |
| ai/gemini | `ai/gemini` | Google Gemini 客户端 | `New`, `WithModel`, `WithEmbeddingModel` |
| ai/conversation | `ai/conversation` | 多轮对话会话管理 | `New`, `Chat`, `ChatStream`, `WithMemory`, `NewBufferMemory`, `NewWindowMemory` |
| ai/embedding | `ai/embedding` | 嵌入向量工具 | `BatchEmbed`, `CosineSimilarity` |
| ai/prompt | `ai/prompt` | 消息模板引擎 | `New`, `MustNew`, `Render`, `MustRender` |
| ai/middleware | `ai/middleware` | AI 模型中间件链 | `Chain`, `Logging`, `Retry`, `RateLimit`, `UsageTracker` |
| ai/vectorstore | `ai/vectorstore` | 向量存储接口 | `VectorStore`, `Document`, `SimilaritySearch`, `WithFilter` |

### 分布式模式 → 详见 `distributed` skill

| 模块 | 包路径 | 描述 | 核心类型/函数 |
|------|--------|------|--------------|
| domain/cqrs | `domain/cqrs` | CQRS 命令/查询模式 | `ChainCommand`, `ApplyCommand`, `ChainQuery`, `ApplyQueryHandler` |
| domain/outbox | `domain/outbox` | Outbox 事务消息 | `NewGORMStore`, `NewRelay`, `InjectTx`, `WithTx` |
| domain | `domain` | 领域事件总线 | `NewEventBus`, `NewAsyncEventBus`, `NewJSONEventConverter` |
| domain/saga | `domain/saga` | Saga 分布式事务编排 | `NewSaga`, `Step`, `Compensate`, `Execute` |

### 消息与任务 → 详见 `pubsub` skill

| 模块 | 包路径 | 描述 | 核心类型/函数 |
|------|--------|------|--------------|
| pubsub | `messaging/pubsub` | 统一 Pub/Sub 接口 | `Publisher`, `Subscriber`, `Message` |
| pubsub/kafka | `messaging/pubsub/kafka` | Kafka driver | `NewPublisher`, `NewSubscriber` |
| pubsub/rabbitmq | `messaging/pubsub/rabbitmq` | RabbitMQ driver | `NewPublisher`, `NewSubscriber` |
| pubsub/redis | `messaging/pubsub/redis` | Redis Streams driver | `NewPublisher`, `NewSubscriber` |
| jobqueue | `messaging/jobqueue` | 异步任务队列 | `NewClient`, `NewWorker`, `Store` |
| jobqueue/redis | `messaging/jobqueue/redis` | Redis Store | `NewStore` |
| jobqueue/kafka | `messaging/jobqueue/kafka` | Kafka Store | `NewStore` |
| jobqueue/rabbitmq | `messaging/jobqueue/rabbitmq` | RabbitMQ Store | `NewStore` |
| jobqueue/database | `messaging/jobqueue/database` | GORM Database Store | `NewStore` |

### 通知

| 模块 | 包路径 | 描述 | 核心类型/函数 |
|------|--------|------|--------------|
| notify | `notify` | 通知发送统一接口 | `Sender`, `Message`, `SendOption` |
| notify/email | `notify/email` | 邮件通知 | `NewSender`, `WithSMTP`, `WithFrom` |
| notify/sms | `notify/sms` | 短信通知 | `NewSender`, `WithProvider` |
| notify/push | `notify/push` | 推送通知（APNs/FCM） | `NewSender`, `WithAPNs`, `WithFCM` |
| notify/nwebhook | `notify/nwebhook` | Webhook 通知发送 | `NewSender`, `WithURL`, `WithSigner` |
| notify/webhook | `notify/webhook` | Webhook 投递与接收 | `NewDispatcher`, `NewReceiver`, `NewHMACSigner` |
| notify/webhook/store/memory | `notify/webhook/store/memory` | 内存 SubscriptionStore | `NewStore` |
| notify/webhook/store/gorm | `notify/webhook/store/gorm` | GORM SubscriptionStore | `NewStore` |

### OAuth2 → 详见 `oauth2` skill

| 模块 | 包路径 | 描述 | 核心类型/函数 |
|------|--------|------|--------------|
| oauth2 | `oauth2` | Provider/StateStore 接口 | `Provider`, `StateStore`, `Token`, `UserInfo` |
| oauth2/state | `oauth2/state` | State 管理 | `NewMemoryStore`, `NewRedisStore` |
| oauth2/github | `oauth2/github` | GitHub OAuth2 | `NewProvider` |
| oauth2/google | `oauth2/google` | Google OAuth2 | `NewProvider` |
| oauth2/wechat | `oauth2/wechat` | 微信 OAuth2 | `NewProvider` |

### OpenAPI → 详见 `openapi` skill

| 模块 | 包路径 | 描述 | 核心类型/函数 |
|------|--------|------|--------------|
| openapi | `openapi` | Code-first OpenAPI 3.0 生成 | `NewRegistry`, `SchemaFrom`, `GET/POST/PUT/DELETE/PATCH`, `ServeJSON`, `ServeYAML` |

### 请求上下文

| 模块 | 包路径 | 描述 | 核心类型/函数 |
|------|--------|------|--------------|
| httpx | `httpx` | HTTP 请求上下文工具（clientip/useragent/botdetect/locale 等） | `HTTPMiddleware`, `GRPCInterceptor` |

### 错误处理

| 模块 | 包路径 | 描述 | 核心类型/函数 |
|------|--------|------|--------------|
| errors | `errors` | 统一业务错误（含 HTTP/gRPC 状态映射） | `New`, `Error`, `WithHTTP`, `WithGRPC`, `FromHTTPStatus`, `FromGRPCCode` |

### 国际化

| 模块 | 包路径 | 描述 | 核心类型/函数 |
|------|--------|------|--------------|
| i18n | `i18n` | 国际化本地化 | `NewBundle`, `LoadFiles`, `Translate`, `WithLogger` |

### 多租户

| 模块 | 包路径 | 描述 | 核心类型/函数 |
|------|--------|------|--------------|
| tenant | `tenant` | 多租户支持 | `Tenant`, `WithTenant`, `FromContext`, `ID`, `HTTPMiddleware`, `Resolver`, `KeySpace` |

### 工具库

| 模块 | 包路径 | 描述 | 核心类型/函数 |
|------|--------|------|--------------|
| xutil/pagination | `xutil/pagination` | 分页工具 | `Params`, `Result`, `Paginate` |
| xutil/crypto | `xutil/crypto` | 加密工具 | `HashPassword`, `ComparePassword` |
| xutil/ptrx | `xutil/ptrx` | 指针工具 | `Of`, `Value`, `ValueOr` |
| xutil/optionx | `xutil/optionx` | 选项模式工具 | `Option`, `Apply` |
| xutil/version | `xutil/version` | 版本信息 | `Version`, `Print` |
| xutil/sorting | `xutil/sorting` | 排序工具 | `Sort`, `GORMScope` |
| xutil/copier | `xutil/copier` | 结构体拷贝 | `Copy`, `CopySlice` |
| xutil/syncx | `xutil/syncx` | 并发工具 | `Map`, `Pool`, `LimitPool`, `SegmentKeysLock` |
| xutil/strx | `xutil/strx` | 字符串工具 | 字符串处理函数集 |
| xutil/randx | `xutil/randx` | 随机数工具 | 随机数/字符串生成 |
| xutil/iox | `xutil/iox` | IO 工具 | IO 辅助函数 |
| xutil/valuex | `xutil/valuex` | 值工具 | `AnyValue` 类型安全取值 |

## 维护说明

当 servex 公开 API 发生变更时：

1. 更新本文件对应模块的索引表格
2. 在对应分组 skill 文件（`servex-transport.md` 等）中修正代码示例
3. 在 `docs/superpowers/examples/` 对应目录修正完整示例
4. 运行 `go -C docs/superpowers/examples build ./...` 确认编译通过

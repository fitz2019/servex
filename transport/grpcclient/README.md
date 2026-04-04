# grpcclient

`grpcclient` 包提供 gRPC 客户端封装，集成服务发现机制，简化远程 gRPC 服务的连接和调用。

## 功能特性

- 集成服务发现，通过服务名自动解析目标地址
- 内置 Keepalive 和 WaitForReady 配置
- 支持自定义拦截器和 DialOption
- 自动管理连接生命周期

## 安装

```bash
go get github.com/Tsukikage7/servex/transport/grpcclient
```

## API

### Client

```go
func New(opts ...Option) (*Client, error)
func (c *Client) Conn() *grpc.ClientConn
func (c *Client) Close() error
```

`New` 方法会通过服务发现解析目标地址并建立 gRPC 连接。若 `serviceName`、`discovery` 或 `logger` 未设置，将触发 panic。

### 配置选项

| 选项               | 默认值        | 说明                     |
| ------------------ | ------------- | ------------------------ |
| `WithName`         | `gRPC-Client` | 客户端名称（用于日志）   |
| `WithServiceName`  | -             | 目标服务名称（必需）     |
| `WithDiscovery`    | -             | 服务发现实例（必需）     |
| `WithLogger`       | -             | 日志记录器（必需）       |
| `WithInterceptors` | -             | 添加一元客户端拦截器     |
| `WithDialOptions`  | -             | 添加额外 gRPC DialOption |

### 服务发现集成

客户端在创建时通过 `discovery.Discovery` 接口解析服务地址：

```go
// 使用 Consul 服务发现
client, err := grpcclient.New(
    grpcclient.WithServiceName("order-service"),
    grpcclient.WithDiscovery(consulDiscovery),
    grpcclient.WithLogger(log),
    grpcclient.WithInterceptors(tracingInterceptor),
)
```

### 错误处理

| 错误                 | 说明                         |
| -------------------- | ---------------------------- |
| `ErrDiscoveryFailed` | 服务发现失败                 |
| `ErrServiceNotFound` | 指定服务名未找到实例         |
| `ErrConnectionFailed`| gRPC 连接建立失败            |

## 许可证

详见项目根目录 LICENSE 文件。

# Discovery 服务发现包

提供微服务架构中的服务注册与发现功能，支持 Consul 作为服务注册中心。

## 功能特性

- 服务注册与注销
- 服务发现
- 多协议支持（HTTP/gRPC）
- 健康检查配置
- 中文错误信息

## 安装

```bash
go get github.com/Tsukikage7/servex/discovery
```

## API 参考

### 类型常量

```go
const (
    TypeConsul = "consul"  // Consul 服务发现
)

const (
    ProtocolHTTP = "http"  // HTTP 协议
    ProtocolGRPC = "grpc"  // gRPC 协议
)
```

### 默认值

| 配置项         | 默认值 |
| -------------- | ------ |
| 健康检查间隔   | 10s    |
| 健康检查超时   | 3s     |
| 失败后注销时间 | 30s    |
| 服务版本       | 1.0.0  |

### 错误类型

| 错误                     | 说明                 |
| ------------------------ | -------------------- |
| `ErrNilConfig`           | 配置为空             |
| `ErrNilLogger`           | 日志记录器为空       |
| `ErrEmptyName`           | 服务名称为空         |
| `ErrEmptyAddress`        | 服务地址为空         |
| `ErrEmptyServiceID`      | 服务ID为空           |
| `ErrUnsupportedType`     | 不支持的服务发现类型 |
| `ErrUnsupportedProtocol` | 不支持的协议类型     |
| `ErrInvalidAddress`      | 无效的地址格式       |
| `ErrInvalidPort`         | 无效的端口号         |
| `ErrNotFound`            | 未发现任何服务实例   |

## 文件结构

```
discovery/
├── discovery.go      # 接口定义和错误常量
├── config.go         # 配置结构体
├── factory.go        # 工厂函数
├── consul.go         # Consul 实现
├── config_test.go    # 配置测试
├── consul_test.go    # Consul 测试
├── factory_test.go   # 工厂测试
├── discovery_test.go # 接口测试
└── README.md         # 文档
```

## 测试

```bash
# 运行测试
go test -v ./discovery/...

# 运行测试并查看覆盖率
go test -v ./discovery/... -cover

# 生成覆盖率报告
go test ./discovery/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

当前测试覆盖率：**87.8%**

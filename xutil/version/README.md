# version

`version` 提供编译时版本信息的注入与查询。

## 功能特性

- 通过 `-ldflags` 在编译时注入版本、Git Commit、构建时间等信息
- `Get()` 获取结构化版本信息
- `Info.String()` 格式化输出版本字符串

## API

### 包级变量

通过 `-ldflags -X` 注入，未注入时使用默认值。

| 变量 | 默认值 | 说明 |
| --- | --- | --- |
| `Version` | `"dev"` | 版本号 |
| `GitCommit` | `"unknown"` | Git 提交哈希 |
| `BuildTime` | `"unknown"` | 构建时间 |
| `GoVersion` | `runtime.Version()` | Go 版本 |

### Info 结构体

```go
type Info struct {
    Version   string `json:"version"`
    GitCommit string `json:"gitCommit"`
    BuildTime string `json:"buildTime"`
    GoVersion string `json:"goVersion"`
}
```

| 函数/方法 | 签名 | 说明 |
| --- | --- | --- |
| `Get` | `Get() Info` | 获取当前版本信息 |
| `String` | `(Info) String() string` | 格式化为 `version=... commit=... built=... go=...` |

## 许可证

Apache-2.0

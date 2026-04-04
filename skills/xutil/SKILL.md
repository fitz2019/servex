---
name: xutil
description: servex 工具库专家。当用户使用 servex 的 ptrx、optionx、valuex、strx、randx、iox、copier、syncx、sorting、pagination、version、crypto 工具包时触发，提供泛型工具函数的完整用法。
---

# servex 工具库

## ptrx -- 指针工具

```go
import "github.com/Tsukikage7/servex/xutil/ptrx"

// 值转指针（常用于 proto/struct 字面量）
p := ptrx.ToPtr(42)       // *int
s := ptrx.ToPtr("hello")  // *string

// 安全解引用（nil 返回零值）
val := ptrx.Value(p)  // 42
val = ptrx.Value[int](nil)  // 0

// 指针值比较
ptrx.Equal(ptrx.ToPtr(1), ptrx.ToPtr(1)) // true
ptrx.Equal(ptrx.ToPtr(1), nil)           // false

// 切片转指针切片
ptrs := ptrx.ToPtrSlice([]int{1, 2, 3}) // []*int
```

## optionx -- 函数选项模式

```go
import "github.com/Tsukikage7/servex/xutil/optionx"

// 定义配置和选项
type Config struct {
    Addr    string
    Timeout time.Duration
}

func WithAddr(addr string) optionx.Option[Config] {
    return func(c *Config) { c.Addr = addr }
}

// 应用选项
var cfg Config
optionx.Apply(&cfg, WithAddr(":8080"))

// 带错误返回的选项
optionx.ApplyErr(&cfg, func(c *Config) error {
    if c.Addr == "" {
        return errors.New("addr required")
    }
    return nil
})
```

## valuex -- 类型转换

```go
import "github.com/Tsukikage7/servex/xutil/valuex"

// 包装任意值
av := valuex.Of(42)

// 精确类型断言
n, err := av.Int()       // 42, nil
s, err := av.String()    // error: 类型不匹配

// 宽松类型转换（跨数值类型、string→int 等）
n, err = valuex.Of("123").AsInt()     // 123, nil
n64, err := valuex.Of(3.14).AsInt64() // 3, nil
```

**支持的类型：** Int/Int8/Int16/Int32/Int64, Uint*, Float32/Float64, String, Bool, Bytes
**宽松转换：** AsInt, AsInt64（支持跨数值类型和 string 转换）

## strx -- 字符串工具

```go
import "github.com/Tsukikage7/servex/xutil/strx"

strx.IsEmpty("  ")               // true
strx.IsNotEmpty("hello")         // true
strx.TrimAndLower("  Hello  ")   // "hello"
strx.TrimAndUpper("  Hello  ")   // "HELLO"
strx.ToTitle("hello world")      // "Hello world"
strx.Truncate("很长的字符串", 5)  // "很长..."
strx.DefaultIfEmpty("", "默认值") // "默认值"

// 姓名处理
first, last := strx.SplitName("John Doe") // "John", "Doe"
full := strx.JoinName("John", "Doe")      // "John Doe"

// 零分配转换（注意安全约束）
b := strx.UnsafeToBytes("hello")   // 不可修改返回值
s := strx.UnsafeToString([]byte{}) // 不可再修改原 []byte
```

## randx -- 随机数

```go
import "github.com/Tsukikage7/servex/xutil/randx"

// 高性能随机生成器（基于 PCG）
r := randx.New()

// 安全随机生成器（基于 crypto/rand）
r = randx.NewSecure()

r.RandInt(1, 100)          // [1, 100) 随机整数
r.RandInt64(0, 1000000)    // [0, 1000000) 随机 int64
r.RandString(16)           // 16 位可打印 ASCII
r.RandAlphanumeric(32)     // 32 位 [a-zA-Z0-9]
r.RandAlpha(8)             // 8 位 [a-zA-Z]
r.RandDigits(6)            // 6 位 [0-9]

// 泛型：从切片随机取元素
item, ok := randx.RandElement(r, []string{"a", "b", "c"})

// 无放回采样
sample := randx.Sample(r, items, 3) // 随机取 3 个
```

## iox -- I/O 工具

```go
import "github.com/Tsukikage7/servex/xutil/iox"

// 读取
data, err := iox.ReadAll(reader)         // 读全部字节
text, err := iox.ReadString(reader)      // 读全部为 string
lines, err := iox.ReadLines(reader)      // 按行读取
n, err := iox.Drain(reader)              // 丢弃全部内容

// 写入
iox.WriteString(writer, "hello")

// 资源管理
closer := iox.MultiCloser(db, file, conn)  // 依次关闭多个 Closer
iox.CloseAndLog(file, func(err error) { log.Error(err) })

// 限制读取量
lr := iox.LimitReadCloser(resp.Body, 1<<20) // 最多 1MB
```

## copier -- 结构体拷贝

```go
import "github.com/Tsukikage7/servex/xutil/copier"

type UserDTO struct { Name string; Age int }
type UserVO  struct { Name string; Age int; Extra string }

src := &UserDTO{Name: "张三", Age: 30}

// 创建新对象并拷贝
vo, err := copier.Copy[UserVO](src)

// 拷贝到已有对象
dst := &UserVO{}
err = copier.CopyTo(src, dst)

// 带选项：忽略字段、字段映射
vo, err = copier.CopyWithOptions[UserVO](src,
    copier.IgnoreFields("Age"),
    copier.FieldMapping("Name", "FullName"),
)
```

**匹配规则：** 按字段名匹配，支持类型相同、可赋值、可转换的字段

## syncx -- 并发原语

```go
import "github.com/Tsukikage7/servex/xutil/syncx"

// 泛型对象池
pool := syncx.NewPool(func() *bytes.Buffer { return new(bytes.Buffer) })
buf := pool.Get()
defer pool.Put(buf)

// 带上限的对象池
lp := syncx.NewLimitPool(100, func() *Conn { return newConn() })
conn, ok := lp.Get() // 超过 100 个时返回 false

// 泛型并发安全 Map
var m syncx.Map[string, int]
m.Store("key", 42)
val, ok := m.Load("key")

// 分段键锁（减小锁粒度）
lock := syncx.NewSegmentKeysLock(64)
lock.Lock("user:123")
defer lock.Unlock("user:123")

// 带 context 的条件变量
cond := syncx.NewCond(&sync.Mutex{})
// 等待时可被 ctx 取消
err := cond.Wait(ctx)
cond.Signal()    // 唤醒一个
cond.Broadcast() // 唤醒全部
```

## sorting -- 排序参数

```go
import "github.com/Tsukikage7/servex/xutil/sorting"

// 解析排序字符串
s := sorting.New("created_time:desc,name:asc")
s.String() // "created_time desc, name asc"

// 默认降序
s = sorting.New("created_time") // created_time desc

// 白名单过滤（防注入）
s = sorting.New("name:asc,password:desc").Filter("id", "name", "created_time")
// 只保留 name:asc

// 设置默认排序
s = sorting.New("").WithDefault("created_time:desc")

// GORM 集成
db.Scopes(s.GORMScope()).Find(&users)
// 或
s.Apply(db).Find(&users)
```

## pagination -- 分页

```go
import "github.com/Tsukikage7/servex/xutil/pagination"

// 创建分页参数（自动校验边界）
p := pagination.New(1, 20) // page=1, pageSize=20
p.Offset() // 0
p.Limit()  // 20

// 边界保护
p = pagination.New(0, 200)  // page→1, pageSize→100(最大值)

// 分页结果
result := pagination.NewResult(users, 100, p)
result.TotalPages() // 5
result.HasNext()     // true
result.HasPrev()     // false
```

**默认值：** Page=1, PageSize=20, MaxPageSize=100

## version -- 版本信息

```go
import "github.com/Tsukikage7/servex/xutil/version"

// 获取编译时注入的版本信息
info := version.Get()
fmt.Println(info.Version)   // "v1.0.0"
fmt.Println(info.GitCommit) // "abc1234"
fmt.Println(info.String())  // "version=v1.0.0 commit=abc1234 built=2024-01-01 go=go1.22"
```

**编译注入：**
```bash
go build -ldflags "-X github.com/Tsukikage7/servex/xutil/version.Version=v1.0.0 \
  -X github.com/Tsukikage7/servex/xutil/version.GitCommit=$(git rev-parse --short HEAD) \
  -X github.com/Tsukikage7/servex/xutil/version.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
```

## crypto -- 密码与随机 ID

```go
import "github.com/Tsukikage7/servex/xutil/crypto"

// 随机 ID
id, err := crypto.GenerateID() // 32 位十六进制 "a1b2c3d4..."

// 验证码
code := crypto.GenerateVerificationCode() // "042857"

// 业务 ID
bizID := crypto.GenerateBusinessID()     // 9 位随机数 int32
bizID64 := crypto.GenerateBusinessID64() // 18 位随机数 int64

// 随机数范围
n, err := crypto.GenerateRandomInt32(1, 100)  // [1, 100]
n64, err := crypto.GenerateRandomInt64(1, 100)

// 密码哈希（bcrypt）
hashed, err := crypto.HashPassword("mypassword")
err = crypto.VerifyPassword(hashed, "mypassword") // nil = 匹配
```

# bizx/captcha — 验证码管理

提供验证码完整生命周期管理：**生成**、**验证**、**防暴力破解**（最大尝试次数）和**防刷控制**（发送冷却时间）。

## 实现

| 存储构造函数 | 说明 |
|-------------|------|
| `NewMemoryStore()` | 内存存储，适合测试 |
| `NewRedisStore(client)` | Redis 存储，生产推荐 |

## 接口

```go
type Manager interface {
    Generate(ctx, key string) (*Code, error) // 生成验证码
    Verify(ctx, key, code string) error      // 验证（成功后自动删除）
    Invalidate(ctx, key string) error        // 手动使验证码失效
}

type Code struct {
    Key       string    // 验证码标识（如手机号、邮箱）
    Code      string    // 验证码内容
    ExpiresAt time.Time // 过期时间
}
```

## 选项

| 选项 | 默认值 | 说明 |
|------|--------|------|
| `WithLength(n)` | 6 | 验证码长度 |
| `WithExpiration(d)` | 5m | 过期时间 |
| `WithMaxAttempts(n)` | 5 | 最大验证次数（超出自动删除） |
| `WithCooldown(d)` | 60s | 发送冷却时间（防刷） |
| `WithAlphabet(s)` | `"0123456789"` | 字符集（纯数字） |

## 快速上手

```go
store := captcha.NewRedisStore(redisClient)
mgr := captcha.NewManager(store,
    captcha.WithLength(6),
    captcha.WithExpiration(10*time.Minute),
    captcha.WithMaxAttempts(5),
    captcha.WithCooldown(60*time.Second),
)

// 生成验证码（手机号作为 key）
code, err := mgr.Generate(ctx, "+8613800138000")
if errors.Is(err, captcha.ErrCooldown) {
    // 60 秒内不能重复发送
    return errors.New("请稍后再试")
}
// 将 code.Code 通过短信发送给用户

// 验证
err = mgr.Verify(ctx, "+8613800138000", userInputCode)
switch {
case errors.Is(err, captcha.ErrCodeExpired):
    return errors.New("验证码已过期")
case errors.Is(err, captcha.ErrCodeInvalid):
    return errors.New("验证码错误")
case errors.Is(err, captcha.ErrTooManyAttempts):
    return errors.New("验证失败次数过多，请重新获取")
}
// 验证成功，验证码已自动删除
```

## 字母数字验证码

```go
mgr := captcha.NewManager(store,
    captcha.WithLength(8),
    captcha.WithAlphabet("ABCDEFGHJKLMNPQRSTUVWXYZ23456789"), // 去除易混淆字符
)
```

## 错误

| 错误 | 说明 |
|------|------|
| `ErrCodeExpired` | 验证码不存在或已过期 |
| `ErrCodeInvalid` | 验证码错误 |
| `ErrTooManyAttempts` | 超出最大验证次数 |
| `ErrCooldown` | 发送冷却中，请稍后再试 |

---
name: auth
description: servex 认证模块专家。当用户使用 servex 的 auth/jwt、auth/apikey 认证时触发。
---

# servex 认证

## auth/jwt — JWT 签发与验证

```go
// 创建 JWT 服务（缺少 WithLogger 会 panic）
jwtSrv := jwt.NewJWT(
    jwt.WithLogger(log),
    jwt.WithSecretKey("your-secret-key"),
    jwt.WithIssuer("my-service"),
    jwt.WithAccessDuration(2 * time.Hour),
    jwt.WithRefreshDuration(7 * 24 * time.Hour),
)

// 签发令牌
claims := &jwt.StandardClaims{
    RegisteredClaims: gojwt.RegisteredClaims{
        Subject: "user-123",
    },
}
tokenStr, err := jwtSrv.Generate(claims)

// 验证令牌
parsed, err := jwtSrv.Validate(tokenStr)
sub, _ := parsed.GetSubject()
```

完整示例：`docs/superpowers/examples/jwt/main.go`

**与 httpserver 集成：**

```go
// NewAuthenticator 将 JWT 服务包装为 auth.Authenticator 接口
authenticator := jwt.NewAuthenticator(jwtSrv)

srv := httpserver.New(mux,
    httpserver.WithLogger(log),
    httpserver.WithAuth(authenticator, "/api/login", "/healthz"), // 白名单路径无需认证
)
```

**关键类型：**
- `jwt.StandardClaims` — 标准 claims 结构（嵌入 `gojwt.RegisteredClaims`）
- `auth.Principal` — 认证后的用户信息，含 `ID`（不是 `UserID`）

## auth/apikey — API Key 验证

```go
// StaticValidator：硬编码 key 列表（适合内部服务、测试）
validator := apikey.StaticValidator(map[string]string{
    "key-abc": "service-a",
    "key-xyz": "service-b",
})

// CacheValidator：带缓存的动态验证（适合从数据库查询 key）
validator := apikey.CacheValidator(
    func(ctx context.Context, key string) (string, error) {
        // 返回 subject（用户ID/服务名），查不到返回 error
        return db.LookupAPIKey(ctx, key)
    },
    5*time.Minute, // 缓存 TTL
)

// 包装为 Authenticator 接口
authenticator := apikey.New(validator)

// 集成到 httpserver（从 X-API-Key header 读取）
srv := httpserver.New(mux,
    httpserver.WithAuth(authenticator, "/healthz"),
)
```

**关键选项：**
- `apikey.New(validator)` — 构造 `*Authenticator`，不是 `NewAuthenticator`
- `StaticValidator` — 返回 `Validator` 函数类型
- `CacheValidator(lookupFn, ttl)` — 带内存缓存的动态验证

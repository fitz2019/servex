# bizx/locking — 业务锁

提供业务级高级锁，区别于 `storage/lock` 的基础分布式锁，本包额外支持**可重入锁**、**读写锁**、**自动续期**和便利的 `WithLock` 辅助函数。

## 实现

基于 `storage/lock.Locker` 接口构建，支持任意底层分布式锁实现（Redis、etcd 等）。

## 接口

```go
type Lock interface {
    Lock(ctx) error
    Unlock(ctx) error
    Extend(ctx, ttl Duration) error // 续期
}

type ReentrantLock interface {
    Lock
    LockCount() int // 当前重入次数
}

type RWLock interface {
    RLock(ctx) error
    RUnlock(ctx) error
    Lock(ctx) error
    Unlock(ctx) error
}
```

## 构造函数

| 函数 | 说明 |
|------|------|
| `NewLock(locker, key, opts...)` | 创建普通分布式锁（带重试） |
| `NewReentrantLock(locker, key, opts...)` | 创建可重入锁 |
| `NewRWLock(locker, key, opts...)` | 创建读写锁 |

## 选项

| 选项 | 默认值 | 说明 |
|------|--------|------|
| `WithTTL(d)` | 30s | 锁持有超时 |
| `WithRetryInterval(d)` | 100ms | 抢锁重试间隔 |
| `WithRetryTimeout(d)` | 10s | 抢锁总超时 |

## 快速上手

```go
locker, _ := storagelock.NewLocker(redisClient)

// 普通锁（推荐使用 WithLock 辅助）
l := locking.NewLock(locker, "order:123", locking.WithTTL(30*time.Second))
err := locking.WithLock(ctx, l, func(ctx context.Context) error {
    return processOrder(ctx, 123)
})

// 可重入锁（同一流程中多次调用同一锁）
rl := locking.NewReentrantLock(locker, "resource:abc")
rl.Lock(ctx)
rl.Lock(ctx) // 不阻塞，计数 +1
rl.Unlock(ctx)
rl.Unlock(ctx) // 计数归零，真正释放

// 读写锁
rwl := locking.NewRWLock(locker, "config")
locking.WithRLock(ctx, rwl, func(ctx context.Context) error {
    return readConfig(ctx)
})
```

## 错误

| 错误 | 说明 |
|------|------|
| `ErrLockFailed` | 在重试超时内未能获取锁 |
| `ErrNotLocked` | 尝试解锁但当前未持有锁 |
| `ErrLockExpired` | 锁已过期 |

// Package locking 提供业务级锁实现.
// 区别于 storage/lock 的基础分布式锁，本包提供可重入锁、读写锁、
// 自动续期等高级锁能力.
// 基本用法:
//	// 普通锁
//	l := locking.NewLock(locker, "order:123", locking.WithTTL(30*time.Second))
//	err := locking.WithLock(ctx, l, func(ctx context.Context) error {
//	    return processOrder(123)
//	})
//	// 可重入锁
//	rl := locking.NewReentrantLock(locker, "resource:abc")
//	rl.Lock(ctx)
//	rl.Lock(ctx) // 同一 goroutine 可再次获取
//	rl.Unlock(ctx)
//	rl.Unlock(ctx)
//	// 读写锁
//	rwl := locking.NewRWLock(locker, "config")
//	locking.WithRLock(ctx, rwl, func(ctx context.Context) error {
//	    return readConfig()
//	})
package locking

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	storagelock "github.com/Tsukikage7/servex/storage/lock"
)

var (
	// ErrLockFailed 获取锁失败.
	ErrLockFailed = errors.New("locking: failed to acquire lock")
	// ErrNotLocked 未持有锁.
	ErrNotLocked = errors.New("locking: not locked")
	// ErrLockExpired 锁已过期.
	ErrLockExpired = errors.New("locking: lock expired")
)

// Lock 锁接口.
type Lock interface {
	Lock(ctx context.Context) error
	Unlock(ctx context.Context) error
	Extend(ctx context.Context, ttl time.Duration) error
}

// ReentrantLock 可重入锁接口.
type ReentrantLock interface {
	Lock
	LockCount() int
}

// RWLock 读写锁接口.
type RWLock interface {
	RLock(ctx context.Context) error
	RUnlock(ctx context.Context) error
	Lock(ctx context.Context) error
	Unlock(ctx context.Context) error
}

// options 锁选项.
type options struct {
	ttl           time.Duration
	retryInterval time.Duration
	retryTimeout  time.Duration
}

// Option 锁选项函数.
type Option func(*options)

// WithTTL 设置锁的过期时间，默认 30s.
func WithTTL(ttl time.Duration) Option {
	return func(o *options) {
		o.ttl = ttl
	}
}

// WithRetryInterval 设置重试间隔，默认 100ms.
func WithRetryInterval(d time.Duration) Option {
	return func(o *options) {
		o.retryInterval = d
	}
}

// WithRetryTimeout 设置重试超时，默认 10s.
func WithRetryTimeout(d time.Duration) Option {
	return func(o *options) {
		o.retryTimeout = d
	}
}

func applyOptions(opts []Option) *options {
	o := &options{
		ttl:           30 * time.Second,
		retryInterval: 100 * time.Millisecond,
		retryTimeout:  10 * time.Second,
	}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

// WithLock 在锁保护下执行函数.
func WithLock(ctx context.Context, lock Lock, fn func(ctx context.Context) error) error {
	if err := lock.Lock(ctx); err != nil {
		return err
	}
	defer lock.Unlock(ctx) //nolint:errcheck
	return fn(ctx)
}

// WithRLock 在读锁保护下执行函数.
func WithRLock(ctx context.Context, rwlock RWLock, fn func(ctx context.Context) error) error {
	if err := rwlock.RLock(ctx); err != nil {
		return err
	}
	defer rwlock.RUnlock(ctx) //nolint:errcheck
	return fn(ctx)
}

// ---- 普通锁实现 ----

// simpleLock 普通分布式锁，包装 storage/lock.Locker 并添加重试逻辑.
type simpleLock struct {
	locker storagelock.Locker
	key    string
	opts   *options
}

// NewLock 创建普通分布式锁.
func NewLock(locker storagelock.Locker, key string, opts ...Option) Lock {
	return &simpleLock{
		locker: locker,
		key:    key,
		opts:   applyOptions(opts),
	}
}

func (l *simpleLock) Lock(ctx context.Context) error {
	deadline := time.Now().Add(l.opts.retryTimeout)
	for {
		acquired, err := l.locker.TryLock(ctx, l.key, l.opts.ttl)
		if err != nil {
			return err
		}
		if acquired {
			return nil
		}
		if time.Now().After(deadline) {
			return ErrLockFailed
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(l.opts.retryInterval):
		}
	}
}

func (l *simpleLock) Unlock(ctx context.Context) error {
	return l.locker.Unlock(ctx, l.key)
}

func (l *simpleLock) Extend(ctx context.Context, ttl time.Duration) error {
	return l.locker.Extend(ctx, l.key, ttl)
}

// ---- 可重入锁实现 ----

// reentrantLock 可重入锁，同一 goroutine 可多次获取.
type reentrantLock struct {
	locker storagelock.Locker
	key    string
	opts   *options

	mu    sync.Mutex
	count int32
	owner int64 // 持有锁的标识（使用 atomic 自增 ID）
}

// 全局计数器，用于生成锁持有者标识.
var lockOwnerID atomic.Int64

// NewReentrantLock 创建可重入锁.
func NewReentrantLock(locker storagelock.Locker, key string, opts ...Option) ReentrantLock {
	return &reentrantLock{
		locker: locker,
		key:    key,
		opts:   applyOptions(opts),
	}
}

func (l *reentrantLock) Lock(ctx context.Context) error {
	l.mu.Lock()
	if l.count > 0 {
		l.count++
		l.mu.Unlock()
		return nil
	}
	l.mu.Unlock()

	// 尝试获取底层锁
	deadline := time.Now().Add(l.opts.retryTimeout)
	for {
		acquired, err := l.locker.TryLock(ctx, l.key, l.opts.ttl)
		if err != nil {
			return err
		}
		if acquired {
			l.mu.Lock()
			l.count = 1
			l.owner = lockOwnerID.Add(1)
			l.mu.Unlock()
			return nil
		}
		if time.Now().After(deadline) {
			return ErrLockFailed
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(l.opts.retryInterval):
		}
	}
}

func (l *reentrantLock) Unlock(ctx context.Context) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.count <= 0 {
		return ErrNotLocked
	}
	l.count--
	if l.count == 0 {
		return l.locker.Unlock(ctx, l.key)
	}
	return nil
}

func (l *reentrantLock) Extend(ctx context.Context, ttl time.Duration) error {
	l.mu.Lock()
	if l.count <= 0 {
		l.mu.Unlock()
		return ErrNotLocked
	}
	l.mu.Unlock()
	return l.locker.Extend(ctx, l.key, ttl)
}

func (l *reentrantLock) LockCount() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return int(l.count)
}

// ---- 读写锁实现 ----

// rwLock 读写锁，使用 storage/lock.Locker 实现.
type rwLock struct {
	locker storagelock.Locker
	key    string
	opts   *options

	mu      sync.Mutex
	readers int32 // 当前读者数量
}

// NewRWLock 创建读写锁.
func NewRWLock(locker storagelock.Locker, key string, opts ...Option) RWLock {
	return &rwLock{
		locker: locker,
		key:    key,
		opts:   applyOptions(opts),
	}
}

func (l *rwLock) writerKey() string {
	return l.key + ":w"
}

func (l *rwLock) readerKey() string {
	return l.key + ":r"
}

func (l *rwLock) RLock(ctx context.Context) error {
	// 获取读锁：先确保没有写锁
	deadline := time.Now().Add(l.opts.retryTimeout)
	for {
		// 检查写锁是否被持有
		acquired, err := l.locker.TryLock(ctx, l.writerKey(), l.opts.ttl)
		if err != nil {
			return err
		}
		if acquired {
			// 写锁没被持有，释放它，增加读者计数
			_ = l.locker.Unlock(ctx, l.writerKey())
			l.mu.Lock()
			l.readers++
			l.mu.Unlock()
			return nil
		}
		if time.Now().After(deadline) {
			return ErrLockFailed
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(l.opts.retryInterval):
		}
	}
}

func (l *rwLock) RUnlock(_ context.Context) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.readers <= 0 {
		return ErrNotLocked
	}
	l.readers--
	return nil
}

func (l *rwLock) Lock(ctx context.Context) error {
	deadline := time.Now().Add(l.opts.retryTimeout)
	for {
		acquired, err := l.locker.TryLock(ctx, l.writerKey(), l.opts.ttl)
		if err != nil {
			return err
		}
		if acquired {
			// 等待所有读者释放
			l.mu.Lock()
			readers := l.readers
			l.mu.Unlock()
			if readers == 0 {
				return nil
			}
			// 还有读者，释放写锁并重试
			_ = l.locker.Unlock(ctx, l.writerKey())
		}
		if time.Now().After(deadline) {
			return ErrLockFailed
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(l.opts.retryInterval):
		}
	}
}

func (l *rwLock) Unlock(ctx context.Context) error {
	return l.locker.Unlock(ctx, l.writerKey())
}

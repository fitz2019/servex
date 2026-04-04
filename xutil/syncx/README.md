# syncx

`syncx` 提供泛型并发工具集，包装标准库并发原语以提供类型安全。

## 功能特性

- `Map[K, V]` -- 泛型并发安全 Map，零值可用
- `Pool[T]` -- 泛型对象池
- `LimitPool[T]` -- 带容量限制的对象池，超过上限时 Get 返回 false
- `SegmentKeysLock` -- 分段键锁，通过 FNV-1a 哈希分段减小锁粒度

## API

### Map[K comparable, V any]

零值可用，无需初始化。

| 方法 | 说明 |
| --- | --- |
| `Load(key K) (V, bool)` | 加载值 |
| `Store(key K, value V)` | 存储值 |
| `LoadOrStore(key K, value V) (V, bool)` | 加载或存储 |
| `LoadAndDelete(key K) (V, bool)` | 加载并删除 |
| `Delete(key K)` | 删除 |
| `Range(fn func(K, V) bool)` | 遍历 |

### Pool[T any]

| 函数/方法 | 说明 |
| --- | --- |
| `NewPool[T](factory func() T) *Pool[T]` | 创建对象池 |
| `Get() T` | 获取对象 |
| `Put(t T)` | 归还对象 |

### LimitPool[T any]

| 函数/方法 | 说明 |
| --- | --- |
| `NewLimitPool[T](maxTokens int, factory func() T) *LimitPool[T]` | 创建带容量限制的对象池 |
| `Get() (T, bool)` | 获取对象，达到上限返回 `(零值, false)` |
| `Put(t T)` | 归还对象并释放令牌 |

### SegmentKeysLock

建议 size 使用 2 的幂次以获得更均匀的哈希分布。

| 函数/方法 | 说明 |
| --- | --- |
| `NewSegmentKeysLock(size uint32) *SegmentKeysLock` | 创建分段键锁 |
| `Lock(key string)` | 写锁定 |
| `TryLock(key string) bool` | 尝试写锁定 |
| `Unlock(key string)` | 写解锁 |
| `RLock(key string)` | 读锁定 |
| `TryRLock(key string) bool` | 尝试读锁定 |
| `RUnlock(key string)` | 读解锁 |

## 许可证

Apache-2.0

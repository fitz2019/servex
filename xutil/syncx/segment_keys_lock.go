package syncx

import "sync"

// SegmentKeysLock 分段键锁.
// 通过对 key 进行哈希分段，减小锁粒度提升并发性能.
type SegmentKeysLock struct {
	locks []sync.RWMutex
	size  uint32
}

// NewSegmentKeysLock 创建分段键锁，size 建议使用 2 的幂次.
func NewSegmentKeysLock(size uint32) *SegmentKeysLock {
	if size == 0 {
		size = 16
	}
	return &SegmentKeysLock{
		locks: make([]sync.RWMutex, size),
		size:  size,
	}
}

func (s *SegmentKeysLock) Lock(key string)        { s.locks[s.hash(key)].Lock() }
func (s *SegmentKeysLock) TryLock(key string) bool { return s.locks[s.hash(key)].TryLock() }
func (s *SegmentKeysLock) Unlock(key string)       { s.locks[s.hash(key)].Unlock() }
func (s *SegmentKeysLock) RLock(key string)        { s.locks[s.hash(key)].RLock() }
func (s *SegmentKeysLock) TryRLock(key string) bool { return s.locks[s.hash(key)].TryRLock() }
func (s *SegmentKeysLock) RUnlock(key string)       { s.locks[s.hash(key)].RUnlock() }

// hash FNV-1a 零分配实现.
func (s *SegmentKeysLock) hash(key string) uint32 {
	const (
		offset32 = uint32(2166136261)
		prime32  = uint32(16777619)
	)
	h := offset32
	for i := range len(key) {
		h ^= uint32(key[i])
		h *= prime32
	}
	return h % s.size
}

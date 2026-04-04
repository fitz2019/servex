package syncx

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/suite"
)

type SegmentKeysLockTestSuite struct {
	suite.Suite
}

func TestSegmentKeysLockSuite(t *testing.T) {
	suite.Run(t, new(SegmentKeysLockTestSuite))
}

func (s *SegmentKeysLockTestSuite) TestNewSegmentKeysLock() {
	skl := NewSegmentKeysLock(16)
	s.NotNil(skl)
	s.Equal(uint32(16), skl.size)
}

func (s *SegmentKeysLockTestSuite) TestNewSegmentKeysLock_ZeroSize() {
	skl := NewSegmentKeysLock(0)
	s.Equal(uint32(16), skl.size)
}

func (s *SegmentKeysLockTestSuite) TestLockUnlock() {
	skl := NewSegmentKeysLock(8)
	skl.Lock("user:1")
	skl.Unlock("user:1")
}

func (s *SegmentKeysLockTestSuite) TestTryLock() {
	skl := NewSegmentKeysLock(8)

	s.True(skl.TryLock("key"))
	s.False(skl.TryLock("key"))
	skl.Unlock("key")

	s.True(skl.TryLock("key"))
	skl.Unlock("key")
}

func (s *SegmentKeysLockTestSuite) TestRLockRUnlock() {
	skl := NewSegmentKeysLock(8)

	skl.RLock("key")
	skl.RLock("key")
	skl.RUnlock("key")
	skl.RUnlock("key")
}

func (s *SegmentKeysLockTestSuite) TestTryRLock() {
	skl := NewSegmentKeysLock(8)

	s.True(skl.TryRLock("key"))
	s.True(skl.TryRLock("key"))
	skl.RUnlock("key")
	skl.RUnlock("key")
}

func (s *SegmentKeysLockTestSuite) TestTryRLock_BlockedByWriteLock() {
	skl := NewSegmentKeysLock(8)

	skl.Lock("key")
	s.False(skl.TryRLock("key"))
	skl.Unlock("key")
}

func (s *SegmentKeysLockTestSuite) TestDifferentKeysIndependent() {
	skl := NewSegmentKeysLock(1024) // 大分段减少冲突
	skl.Lock("key1")

	s.True(skl.TryLock("key2"))
	skl.Unlock("key2")
	skl.Unlock("key1")
}

func (s *SegmentKeysLockTestSuite) TestConcurrentDifferentKeys() {
	skl := NewSegmentKeysLock(32)
	counter := make(map[string]int)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for i := range 100 {
		key := fmt.Sprintf("key:%d", i%10)
		wg.Go(func() {
			skl.Lock(key)
			defer skl.Unlock(key)

			mu.Lock()
			counter[key]++
			mu.Unlock()
		})
	}
	wg.Wait()

	total := 0
	for _, v := range counter {
		total += v
	}
	s.Equal(100, total)
}

func (s *SegmentKeysLockTestSuite) TestConcurrentReadWrite() {
	skl := NewSegmentKeysLock(16)
	var wg sync.WaitGroup

	for i := range 50 {
		key := fmt.Sprintf("item:%d", i%5)

		wg.Go(func() {
			skl.Lock(key)
			skl.Unlock(key)
		})
		wg.Go(func() {
			skl.RLock(key)
			skl.RUnlock(key)
		})
	}
	wg.Wait()
}

func (s *SegmentKeysLockTestSuite) TestHashDistribution() {
	skl := NewSegmentKeysLock(8)

	indices := make(map[uint32]bool)
	for i := range 100 {
		idx := skl.hash(fmt.Sprintf("key-%d", i))
		indices[idx] = true
	}
	s.Greater(len(indices), 4)
}

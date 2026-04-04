package syncx

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/suite"
)

type LimitPoolTestSuite struct {
	suite.Suite
}

func TestLimitPoolSuite(t *testing.T) {
	suite.Run(t, new(LimitPoolTestSuite))
}

func (s *LimitPoolTestSuite) TestNewLimitPool() {
	lp := NewLimitPool(5, func() int { return 0 })
	s.NotNil(lp)
}

func (s *LimitPoolTestSuite) TestGetWithinLimit() {
	lp := NewLimitPool(3, func() []byte {
		return make([]byte, 0, 64)
	})

	v1, ok := lp.Get()
	s.True(ok)
	s.NotNil(v1)

	v2, ok := lp.Get()
	s.True(ok)
	s.NotNil(v2)

	v3, ok := lp.Get()
	s.True(ok)
	s.NotNil(v3)
}

func (s *LimitPoolTestSuite) TestGetExceedLimit() {
	lp := NewLimitPool(2, func() int { return 42 })

	_, ok1 := lp.Get()
	s.True(ok1)

	_, ok2 := lp.Get()
	s.True(ok2)

	val, ok3 := lp.Get()
	s.False(ok3)
	s.Equal(0, val)
}

func (s *LimitPoolTestSuite) TestPutReleasesToken() {
	lp := NewLimitPool(1, func() int { return 1 })

	v, ok := lp.Get()
	s.True(ok)

	_, ok = lp.Get()
	s.False(ok)

	lp.Put(v)

	v2, ok := lp.Get()
	s.True(ok)
	s.Equal(1, v2)
}

func (s *LimitPoolTestSuite) TestConcurrentAccess() {
	const maxTokens = 10
	lp := NewLimitPool(maxTokens, func() int { return 0 })

	var wg sync.WaitGroup
	successCount := make(chan struct{}, 1000)

	for range 100 {
		wg.Go(func() {
			for range 10 {
				if val, ok := lp.Get(); ok {
					successCount <- struct{}{}
					lp.Put(val)
				}
			}
		})
	}
	wg.Wait()
	close(successCount)

	count := 0
	for range successCount {
		count++
	}
	s.Greater(count, 0)
}

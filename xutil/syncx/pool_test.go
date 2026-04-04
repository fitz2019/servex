package syncx

import (
	"bytes"
	"sync"
	"testing"

	"github.com/stretchr/testify/suite"
)

type PoolTestSuite struct {
	suite.Suite
}

func TestPoolSuite(t *testing.T) {
	suite.Run(t, new(PoolTestSuite))
}

func (s *PoolTestSuite) TestNewPool() {
	p := NewPool(func() *bytes.Buffer {
		return bytes.NewBuffer(nil)
	})
	s.NotNil(p)
}

func (s *PoolTestSuite) TestGetPut() {
	p := NewPool(func() []byte {
		return make([]byte, 0, 1024)
	})

	buf := p.Get()
	s.NotNil(buf)
	s.Equal(0, len(buf))
	s.Equal(1024, cap(buf))

	buf = append(buf, 1, 2, 3)
	p.Put(buf[:0])
}

func (s *PoolTestSuite) TestGetReturnsNewObject() {
	counter := 0
	p := NewPool(func() int {
		counter++
		return counter
	})

	val := p.Get()
	s.Equal(1, val)
}

func (s *PoolTestSuite) TestConcurrentAccess() {
	p := NewPool(func() *bytes.Buffer {
		return bytes.NewBuffer(nil)
	})

	var wg sync.WaitGroup
	for range 100 {
		wg.Go(func() {
			buf := p.Get()
			buf.WriteString("hello")
			buf.Reset()
			p.Put(buf)
		})
	}
	wg.Wait()
}

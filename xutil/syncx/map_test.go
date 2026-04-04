package syncx

import (
	"fmt"
	"sort"
	"sync"
	"testing"

	"github.com/stretchr/testify/suite"
)

type MapTestSuite struct {
	suite.Suite
}

func TestMapSuite(t *testing.T) {
	suite.Run(t, new(MapTestSuite))
}

func (s *MapTestSuite) TestZeroValue() {
	var m Map[string, int]
	_, ok := m.Load("key")
	s.False(ok)
}

func (s *MapTestSuite) TestStoreAndLoad() {
	var m Map[string, int]
	m.Store("a", 1)
	m.Store("b", 2)

	val, ok := m.Load("a")
	s.True(ok)
	s.Equal(1, val)

	val, ok = m.Load("b")
	s.True(ok)
	s.Equal(2, val)
}

func (s *MapTestSuite) TestLoadMissing() {
	var m Map[string, int]
	val, ok := m.Load("missing")
	s.False(ok)
	s.Equal(0, val)
}

func (s *MapTestSuite) TestLoadOrStore_New() {
	var m Map[string, int]
	val, loaded := m.LoadOrStore("key", 42)
	s.False(loaded)
	s.Equal(42, val)
}

func (s *MapTestSuite) TestLoadOrStore_Existing() {
	var m Map[string, int]
	m.Store("key", 10)

	val, loaded := m.LoadOrStore("key", 42)
	s.True(loaded)
	s.Equal(10, val)
}

func (s *MapTestSuite) TestLoadAndDelete() {
	var m Map[string, int]
	m.Store("key", 100)

	val, loaded := m.LoadAndDelete("key")
	s.True(loaded)
	s.Equal(100, val)

	_, ok := m.Load("key")
	s.False(ok)
}

func (s *MapTestSuite) TestLoadAndDelete_Missing() {
	var m Map[string, int]
	val, loaded := m.LoadAndDelete("key")
	s.False(loaded)
	s.Equal(0, val)
}

func (s *MapTestSuite) TestDelete() {
	var m Map[string, int]
	m.Store("key", 1)
	m.Delete("key")

	_, ok := m.Load("key")
	s.False(ok)
}

func (s *MapTestSuite) TestRange() {
	var m Map[string, int]
	m.Store("a", 1)
	m.Store("b", 2)
	m.Store("c", 3)

	var keys []string
	m.Range(func(key string, value int) bool {
		keys = append(keys, key)
		return true
	})

	sort.Strings(keys)
	s.Equal([]string{"a", "b", "c"}, keys)
}

func (s *MapTestSuite) TestRange_EarlyStop() {
	var m Map[int, string]
	for i := range 10 {
		m.Store(i, "val")
	}

	count := 0
	m.Range(func(key int, value string) bool {
		count++
		return count < 3
	})
	s.Equal(3, count)
}

func (s *MapTestSuite) TestConcurrentAccess() {
	var m Map[int, int]
	var wg sync.WaitGroup

	for i := range 100 {
		wg.Go(func() {
			m.Store(i, i*10)
		})
		wg.Go(func() {
			m.Load(i)
		})
	}
	wg.Wait()
}

func (s *MapTestSuite) TestLoadOrStoreFunc_New() {
	var m Map[string, int]
	val, loaded, err := m.LoadOrStoreFunc("key", func() (int, error) { return 42, nil })
	s.NoError(err)
	s.False(loaded)
	s.Equal(42, val)

	// 再次调用应直接返回已有值
	val2, loaded2, err2 := m.LoadOrStoreFunc("key", func() (int, error) { return 99, nil })
	s.NoError(err2)
	s.True(loaded2)
	s.Equal(42, val2)
}

func (s *MapTestSuite) TestLoadOrStoreFunc_FnError() {
	var m Map[string, int]
	_, loaded, err := m.LoadOrStoreFunc("key", func() (int, error) {
		return 0, fmt.Errorf("compute error")
	})
	s.Error(err)
	s.False(loaded)

	// fn 返回 error 时不应存储
	_, ok := m.Load("key")
	s.False(ok)
}

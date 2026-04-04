package syncx

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type CondTestSuite struct {
	suite.Suite
}

func TestCondSuite(t *testing.T) {
	suite.Run(t, new(CondTestSuite))
}

func (s *CondTestSuite) TestSignalWakesOneWaiter() {
	var mu sync.Mutex
	cond := NewCond(&mu)
	ready := make(chan struct{})

	var woken int
	var wg sync.WaitGroup
	wg.Add(2)

	for range 2 {
		go func() {
			defer wg.Done()
			mu.Lock()
			ready <- struct{}{}
			_ = cond.Wait(s.T().Context())
			woken++
			mu.Unlock()
		}()
	}

	// 等待两个 goroutine 都进入 Wait
	<-ready
	<-ready
	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	cond.Signal()
	mu.Unlock()

	// 再给一次 Signal，让第二个也退出
	time.Sleep(10 * time.Millisecond)
	mu.Lock()
	cond.Signal()
	mu.Unlock()

	wg.Wait()
	s.Equal(2, woken)
}

func (s *CondTestSuite) TestBroadcastWakesAll() {
	var mu sync.Mutex
	cond := NewCond(&mu)
	ready := make(chan struct{}, 3)

	var woken int
	var wg sync.WaitGroup
	wg.Add(3)

	for range 3 {
		go func() {
			defer wg.Done()
			mu.Lock()
			ready <- struct{}{}
			_ = cond.Wait(s.T().Context())
			woken++
			mu.Unlock()
		}()
	}

	for range 3 {
		<-ready
	}
	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	cond.Broadcast()
	mu.Unlock()

	wg.Wait()
	s.Equal(3, woken)
}

func (s *CondTestSuite) TestWaitContextCancel() {
	var mu sync.Mutex
	cond := NewCond(&mu)

	ctx, cancel := context.WithCancel(s.T().Context())

	done := make(chan error, 1)
	go func() {
		mu.Lock()
		err := cond.Wait(ctx)
		mu.Unlock()
		done <- err
	}()

	time.Sleep(10 * time.Millisecond)
	cancel()

	err := <-done
	s.ErrorIs(err, context.Canceled)
}

func (s *CondTestSuite) TestWaitContextTimeout() {
	var mu sync.Mutex
	cond := NewCond(&mu)

	ctx, cancel := context.WithTimeout(s.T().Context(), 20*time.Millisecond)
	defer cancel()

	mu.Lock()
	err := cond.Wait(ctx)
	mu.Unlock()

	s.ErrorIs(err, context.DeadlineExceeded)
}

func (s *CondTestSuite) TestSignalNoWaiter() {
	var mu sync.Mutex
	cond := NewCond(&mu)
	// 无 waiter 时调用 Signal 不应 panic
	s.NotPanics(func() { cond.Signal() })
	s.NotPanics(func() { cond.Broadcast() })
}

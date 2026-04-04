package blockingqueue

import (
	"context"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type BlockingQueueTestSuite struct {
	suite.Suite
}

func TestBlockingQueueSuite(t *testing.T) {
	suite.Run(t, new(BlockingQueueTestSuite))
}

func (s *BlockingQueueTestSuite) TestNew() {
	q := New[int](10)
	s.NotNil(q)
	s.Equal(0, q.Len())
	s.True(q.IsEmpty())
	s.False(q.IsFull())
}

func (s *BlockingQueueTestSuite) TestNew_PanicOnZero() {
	s.Panics(func() { New[int](0) })
}

func (s *BlockingQueueTestSuite) TestNew_PanicOnNegative() {
	s.Panics(func() { New[int](-1) })
}

func (s *BlockingQueueTestSuite) TestEnqueueDequeue() {
	ctx := s.T().Context()
	q := New[int](5)

	s.NoError(q.Enqueue(ctx, 1))
	s.NoError(q.Enqueue(ctx, 2))
	s.NoError(q.Enqueue(ctx, 3))

	s.Equal(3, q.Len())

	val, err := q.Dequeue(ctx)
	s.NoError(err)
	s.Equal(1, val)

	val, err = q.Dequeue(ctx)
	s.NoError(err)
	s.Equal(2, val)

	val, err = q.Dequeue(ctx)
	s.NoError(err)
	s.Equal(3, val)

	s.True(q.IsEmpty())
}

func (s *BlockingQueueTestSuite) TestFIFOOrder() {
	ctx := s.T().Context()
	q := New[int](100)

	for i := range 50 {
		s.NoError(q.Enqueue(ctx, i))
	}

	for i := range 50 {
		val, err := q.Dequeue(ctx)
		s.NoError(err)
		s.Equal(i, val)
	}
}

func (s *BlockingQueueTestSuite) TestIsFull() {
	ctx := s.T().Context()
	q := New[int](2)

	s.NoError(q.Enqueue(ctx, 1))
	s.NoError(q.Enqueue(ctx, 2))
	s.True(q.IsFull())
}

func (s *BlockingQueueTestSuite) TestEnqueueBlocksWhenFull() {
	ctx, cancel := context.WithTimeout(s.T().Context(), 50*time.Millisecond)
	defer cancel()

	q := New[int](1)
	s.NoError(q.Enqueue(ctx, 1))

	err := q.Enqueue(ctx, 2)
	s.Error(err)
	s.ErrorIs(err, context.DeadlineExceeded)
}

func (s *BlockingQueueTestSuite) TestDequeueBlocksWhenEmpty() {
	ctx, cancel := context.WithTimeout(s.T().Context(), 50*time.Millisecond)
	defer cancel()

	q := New[int](5)

	_, err := q.Dequeue(ctx)
	s.Error(err)
	s.ErrorIs(err, context.DeadlineExceeded)
}

func (s *BlockingQueueTestSuite) TestEnqueue_ContextCanceled() {
	ctx, cancel := context.WithCancel(s.T().Context())

	q := New[int](1)
	s.NoError(q.Enqueue(ctx, 1))

	cancel()
	err := q.Enqueue(ctx, 2)
	s.Error(err)
	s.ErrorIs(err, context.Canceled)
}

func (s *BlockingQueueTestSuite) TestDequeue_ContextCanceled() {
	ctx, cancel := context.WithCancel(s.T().Context())
	cancel()

	q := New[int](5)
	_, err := q.Dequeue(ctx)
	s.Error(err)
	s.ErrorIs(err, context.Canceled)
}

func (s *BlockingQueueTestSuite) TestBlockAndResume_Enqueue() {
	ctx := s.T().Context()
	q := New[int](1)
	s.NoError(q.Enqueue(ctx, 1))

	done := make(chan struct{})
	go func() {
		s.NoError(q.Enqueue(ctx, 2))
		close(done)
	}()

	time.Sleep(20 * time.Millisecond)
	val, err := q.Dequeue(ctx)
	s.NoError(err)
	s.Equal(1, val)

	select {
	case <-done:
	case <-time.After(time.Second):
		s.Fail("Enqueue 未在超时前完成")
	}
}

func (s *BlockingQueueTestSuite) TestBlockAndResume_Dequeue() {
	ctx := s.T().Context()
	q := New[int](5)

	done := make(chan int)
	go func() {
		val, err := q.Dequeue(ctx)
		s.NoError(err)
		done <- val
	}()

	time.Sleep(20 * time.Millisecond)
	s.NoError(q.Enqueue(ctx, 42))

	select {
	case val := <-done:
		s.Equal(42, val)
	case <-time.After(time.Second):
		s.Fail("Dequeue 未在超时前完成")
	}
}

func (s *BlockingQueueTestSuite) TestConcurrentProducerConsumer() {
	ctx := s.T().Context()
	q := New[int](10)
	const numProducers = 5
	const numConsumers = 5
	const itemsPerProducer = 100

	var wg sync.WaitGroup

	for i := range numProducers {
		wg.Go(func() {
			for j := range itemsPerProducer {
				s.NoError(q.Enqueue(ctx, i*itemsPerProducer+j))
			}
		})
	}

	results := make(chan int, numProducers*itemsPerProducer)
	for range numConsumers {
		wg.Go(func() {
			for range numProducers * itemsPerProducer / numConsumers {
				val, err := q.Dequeue(ctx)
				s.NoError(err)
				results <- val
			}
		})
	}

	wg.Wait()
	close(results)

	collected := make([]int, 0, numProducers*itemsPerProducer)
	for val := range results {
		collected = append(collected, val)
	}

	sort.Ints(collected)
	s.Equal(numProducers*itemsPerProducer, len(collected))

	expected := make([]int, numProducers*itemsPerProducer)
	for i := range expected {
		expected[i] = i
	}
	s.Equal(expected, collected)
}

func (s *BlockingQueueTestSuite) TestRingBufferWrapAround() {
	ctx := s.T().Context()
	q := New[int](3)

	for round := range 5 {
		for j := range 3 {
			s.NoError(q.Enqueue(ctx, round*10+j))
		}
		for j := range 3 {
			val, err := q.Dequeue(ctx)
			s.NoError(err)
			s.Equal(round*10+j, val)
		}
	}
	s.True(q.IsEmpty())
}

func (s *BlockingQueueTestSuite) TestStringQueue() {
	ctx := s.T().Context()
	q := New[string](5)

	s.NoError(q.Enqueue(ctx, "hello"))
	s.NoError(q.Enqueue(ctx, "world"))

	val, err := q.Dequeue(ctx)
	s.NoError(err)
	s.Equal("hello", val)

	val, err = q.Dequeue(ctx)
	s.NoError(err)
	s.Equal("world", val)
}

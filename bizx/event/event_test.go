package event

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPublish_Subscribe(t *testing.T) {
	b := New()
	defer b.Close()

	var received string
	b.Subscribe("user.created", func(ctx context.Context, evt *Event) error {
		received = evt.Payload.(string)
		return nil
	})

	err := b.Publish(context.Background(), "user.created", "alice")
	require.NoError(t, err)
	assert.Equal(t, "alice", received)
}

func TestWildcard(t *testing.T) {
	b := New()
	defer b.Close()

	var events []string
	var mu sync.Mutex

	// user.* 应匹配 user.created 和 user.deleted
	b.Subscribe("user.*", func(ctx context.Context, evt *Event) error {
		mu.Lock()
		events = append(events, evt.Name)
		mu.Unlock()
		return nil
	})

	_ = b.Publish(context.Background(), "user.created", nil)
	_ = b.Publish(context.Background(), "user.deleted", nil)
	_ = b.Publish(context.Background(), "order.created", nil)        // 不应匹配
	_ = b.Publish(context.Background(), "user.profile.updated", nil) // 不应匹配（多层）

	assert.Len(t, events, 2)
	assert.Contains(t, events, "user.created")
	assert.Contains(t, events, "user.deleted")

	// * 匹配所有
	var allEvents []string
	b.Subscribe("*", func(ctx context.Context, evt *Event) error {
		mu.Lock()
		allEvents = append(allEvents, evt.Name)
		mu.Unlock()
		return nil
	})

	_ = b.Publish(context.Background(), "anything", nil)
	assert.Len(t, allEvents, 1)
}

func TestPriority(t *testing.T) {
	b := New()
	defer b.Close()

	var order []int
	var mu sync.Mutex

	b.Subscribe("test", func(ctx context.Context, evt *Event) error {
		mu.Lock()
		order = append(order, 3)
		mu.Unlock()
		return nil
	}, WithPriority(3))

	b.Subscribe("test", func(ctx context.Context, evt *Event) error {
		mu.Lock()
		order = append(order, 1)
		mu.Unlock()
		return nil
	}, WithPriority(1))

	b.Subscribe("test", func(ctx context.Context, evt *Event) error {
		mu.Lock()
		order = append(order, 2)
		mu.Unlock()
		return nil
	}, WithPriority(2))

	err := b.Publish(context.Background(), "test", nil)
	require.NoError(t, err)

	assert.Equal(t, []int{1, 2, 3}, order)
}

func TestAsync(t *testing.T) {
	b := New(WithBufferSize(100))
	defer b.Close()

	var count atomic.Int32

	b.Subscribe("async-event", func(ctx context.Context, evt *Event) error {
		count.Add(1)
		return nil
	}, WithAsync(true))

	for i := 0; i < 10; i++ {
		err := b.Publish(context.Background(), "async-event", i)
		require.NoError(t, err)
	}

	// 等待异步处理完成
	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, int32(10), count.Load())
}

func TestUnsubscribe(t *testing.T) {
	b := New()
	defer b.Close()

	var called bool
	b.Subscribe("test", func(ctx context.Context, evt *Event) error {
		called = true
		return nil
	})

	b.Unsubscribe("test")

	err := b.Publish(context.Background(), "test", nil)
	require.NoError(t, err)
	assert.False(t, called)
}

func TestClose(t *testing.T) {
	b := New()

	err := b.Close()
	require.NoError(t, err)

	err = b.Publish(context.Background(), "test", nil)
	assert.ErrorIs(t, err, ErrBusClosed)
}

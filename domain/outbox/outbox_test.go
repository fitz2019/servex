package outbox

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/Tsukikage7/servex/messaging/pubsub"
)

// --- mock Publisher ---

type mockPublisher struct {
	mu      sync.Mutex
	sent    []*pubsub.Message
	sendErr error
	closed  bool
}

func newMockPublisher() *mockPublisher {
	return &mockPublisher{}
}

func (p *mockPublisher) Publish(_ context.Context, topic string, msgs ...*pubsub.Message) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.sendErr != nil {
		return p.sendErr
	}
	p.sent = append(p.sent, msgs...)
	return nil
}

func (p *mockPublisher) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.closed = true
	return nil
}

func (p *mockPublisher) sentMessages() []*pubsub.Message {
	p.mu.Lock()
	defer p.mu.Unlock()
	cp := make([]*pubsub.Message, len(p.sent))
	copy(cp, p.sent)
	return cp
}

// --- helpers ---

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	require.NoError(t, err)
	return db
}

func setupTestStore(t *testing.T) (*GORMStore, *gorm.DB) {
	t.Helper()
	db := setupTestDB(t)
	store := NewGORMStoreFromDB(db)
	require.NoError(t, store.AutoMigrate())
	return store, db
}

// --- OutboxMessage 测试 ---

func TestNewOutboxMessage(t *testing.T) {
	msg := &pubsub.Message{
		Topic:   "orders",
		Key:     []byte("order-123"),
		Body:    []byte(`{"id":"123"}`),
		Headers: map[string]string{"trace-id": "abc"},
	}
	om := NewOutboxMessage(msg)

	assert.Equal(t, "orders", om.Topic)
	assert.Equal(t, []byte("order-123"), om.Key)
	assert.Equal(t, []byte(`{"id":"123"}`), om.Value)
	assert.Equal(t, MessageStatus(0), om.Status)

	var h map[string]string
	require.NoError(t, json.Unmarshal([]byte(om.Headers), &h))
	assert.Equal(t, "abc", h["trace-id"])
}

func TestOutboxMessage_ToMessage(t *testing.T) {
	om := &OutboxMessage{
		Topic:   "events",
		Key:     []byte("key-1"),
		Value:   []byte("data"),
		Headers: `{"x":"y"}`,
	}
	msg := om.ToMessage()

	assert.Equal(t, "events", msg.Topic)
	assert.Equal(t, []byte("key-1"), msg.Key)
	assert.Equal(t, []byte("data"), msg.Body)
	assert.Equal(t, "y", msg.Headers["x"])
}

func TestHeadersToJSON_Empty(t *testing.T) {
	assert.Equal(t, "", HeadersToJSON(nil))
	assert.Equal(t, "", HeadersToJSON(map[string]string{}))
}

func TestMessageStatus_String(t *testing.T) {
	assert.Equal(t, "Pending", StatusPending.String())
	assert.Equal(t, "Processing", StatusProcessing.String())
	assert.Equal(t, "Sent", StatusSent.String())
	assert.Equal(t, "Failed", StatusFailed.String())
	assert.Equal(t, "Unknown", MessageStatus(99).String())
}

// --- Store 测试 ---

func TestGORMStore_Save(t *testing.T) {
	store, db := setupTestStore(t)
	ctx := t.Context()

	err := store.Save(ctx,
		&OutboxMessage{Topic: "t1", Value: []byte("v1")},
		&OutboxMessage{Topic: "t2", Value: []byte("v2")},
	)
	require.NoError(t, err)

	var count int64
	db.Model(&OutboxMessage{}).Count(&count)
	assert.Equal(t, int64(2), count)
}

func TestGORMStore_Save_Empty(t *testing.T) {
	store, _ := setupTestStore(t)
	err := store.Save(t.Context())
	assert.NoError(t, err)
}

func TestGORMStore_WithTx(t *testing.T) {
	store, db := setupTestStore(t)
	ctx := t.Context()

	err := store.WithTx(ctx, func(txCtx context.Context) error {
		return store.Save(txCtx,
			&OutboxMessage{Topic: "t1", Value: []byte("v1")},
			&OutboxMessage{Topic: "t2", Value: []byte("v2")},
		)
	})
	require.NoError(t, err)

	var count int64
	db.Model(&OutboxMessage{}).Count(&count)
	assert.Equal(t, int64(2), count)
}

func TestGORMStore_WithTx_Rollback(t *testing.T) {
	store, db := setupTestStore(t)
	ctx := t.Context()

	err := store.WithTx(ctx, func(txCtx context.Context) error {
		if err := store.Save(txCtx, &OutboxMessage{Topic: "t1", Value: []byte("v1")}); err != nil {
			return err
		}
		return errors.New("模拟事务回滚")
	})
	assert.Error(t, err)

	// 事务已回滚，不应有记录
	var count int64
	db.Model(&OutboxMessage{}).Count(&count)
	assert.Equal(t, int64(0), count)
}

func TestInjectTx_ExtractTx(t *testing.T) {
	db := setupTestDB(t)

	ctx := t.Context()
	txCtx := InjectTx(ctx, db)

	tx, ok := ExtractTx(txCtx)
	assert.True(t, ok)
	assert.Equal(t, db, tx)
}

func TestExtractTx_NotInjected(t *testing.T) {
	ctx := t.Context()
	tx, ok := ExtractTx(ctx)
	assert.False(t, ok)
	assert.Nil(t, tx)
}

func TestGORMStore_FetchPending(t *testing.T) {
	store, db := setupTestStore(t)
	ctx := t.Context()

	for i := range 3 {
		db.Create(&OutboxMessage{
			Topic: "topic",
			Value: []byte{byte(i)},
		})
	}

	msgs, err := store.FetchPending(ctx, 2)
	require.NoError(t, err)
	assert.Len(t, msgs, 2)

	var processing int64
	db.Model(&OutboxMessage{}).Where("status = ?", StatusProcessing).Count(&processing)
	assert.Equal(t, int64(2), processing)

	var pending int64
	db.Model(&OutboxMessage{}).Where("status = ?", StatusPending).Count(&pending)
	assert.Equal(t, int64(1), pending)
}

func TestGORMStore_FetchPending_Empty(t *testing.T) {
	store, _ := setupTestStore(t)
	msgs, err := store.FetchPending(t.Context(), 10)
	require.NoError(t, err)
	assert.Empty(t, msgs)
}

func TestGORMStore_MarkSent(t *testing.T) {
	store, db := setupTestStore(t)
	ctx := t.Context()

	db.Create(&OutboxMessage{Topic: "t", Value: []byte("v"), Status: StatusProcessing})
	db.Create(&OutboxMessage{Topic: "t", Value: []byte("v"), Status: StatusProcessing})

	err := store.MarkSent(ctx, []uint64{1, 2})
	require.NoError(t, err)

	var msgs []OutboxMessage
	db.Find(&msgs)
	for _, m := range msgs {
		assert.Equal(t, StatusSent, m.Status)
		assert.NotNil(t, m.SentAt)
	}
}

func TestGORMStore_MarkSent_Empty(t *testing.T) {
	store, _ := setupTestStore(t)
	assert.NoError(t, store.MarkSent(t.Context(), nil))
	assert.NoError(t, store.MarkSent(t.Context(), []uint64{}))
}

func TestGORMStore_MarkFailed(t *testing.T) {
	store, db := setupTestStore(t)
	ctx := t.Context()

	db.Create(&OutboxMessage{Topic: "t", Value: []byte("v"), Status: StatusProcessing})

	err := store.MarkFailed(ctx, 1, "send timeout")
	require.NoError(t, err)

	var msg OutboxMessage
	db.First(&msg, 1)
	assert.Equal(t, StatusFailed, msg.Status)
	assert.Equal(t, 1, msg.RetryCount)
	assert.Equal(t, "send timeout", msg.LastError)
}

func TestGORMStore_ResetStale(t *testing.T) {
	store, db := setupTestStore(t)
	ctx := t.Context()

	db.Create(&OutboxMessage{Topic: "t", Value: []byte("v"), Status: StatusProcessing})
	db.Model(&OutboxMessage{}).Where("id = 1").
		Update("updated_at", time.Now().Add(-10*time.Minute))

	n, err := store.ResetStale(ctx, 5*time.Minute)
	require.NoError(t, err)
	assert.Equal(t, int64(1), n)

	var msg OutboxMessage
	db.First(&msg, 1)
	assert.Equal(t, StatusPending, msg.Status)
}

func TestGORMStore_Cleanup(t *testing.T) {
	store, db := setupTestStore(t)
	ctx := t.Context()

	past := time.Now().Add(-48 * time.Hour)
	db.Create(&OutboxMessage{Topic: "t", Value: []byte("v"), Status: StatusSent, SentAt: &past})

	n, err := store.Cleanup(ctx, time.Now().Add(-24*time.Hour))
	require.NoError(t, err)
	assert.Equal(t, int64(1), n)

	var count int64
	db.Model(&OutboxMessage{}).Count(&count)
	assert.Equal(t, int64(0), count)
}

// --- Relay 测试 ---

func TestNewRelay_NilStore(t *testing.T) {
	_, err := NewRelay(nil, newMockPublisher())
	assert.ErrorIs(t, err, ErrNilStore)
}

func TestNewRelay_NilProducer(t *testing.T) {
	store, _ := setupTestStore(t)
	_, err := NewRelay(store, nil)
	assert.ErrorIs(t, err, ErrNilProducer)
}

func TestRelay_StartStop(t *testing.T) {
	store, _ := setupTestStore(t)
	producer := newMockPublisher()

	relay, err := NewRelay(store, producer,
		WithPollInterval(50*time.Millisecond),
		WithCleanupInterval(50*time.Millisecond),
	)
	require.NoError(t, err)

	ctx := t.Context()
	require.NoError(t, relay.Start(ctx))
	t.Cleanup(func() {
		stopCtx, c := context.WithTimeout(t.Context(), 5*time.Second)
		defer c()
		relay.Stop(stopCtx)
	})

	assert.ErrorIs(t, relay.Start(ctx), ErrRelayAlreadyRunning)

	stopCtx, stopCancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer stopCancel()
	require.NoError(t, relay.Stop(stopCtx))

	assert.ErrorIs(t, relay.Stop(stopCtx), ErrRelayNotRunning)
}

func TestRelay_PollAndDeliver(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping relay test in short mode")
	}
	store, db := setupTestStore(t)
	producer := newMockPublisher()

	relay, err := NewRelay(store, producer,
		WithPollInterval(50*time.Millisecond),
		WithCleanupInterval(time.Hour),
	)
	require.NoError(t, err)

	ctx := t.Context()

	db.Create(&OutboxMessage{
		Topic: "orders",
		Key:   []byte("key-1"),
		Value: []byte(`{"id":"1"}`),
	})

	require.NoError(t, relay.Start(ctx))
	t.Cleanup(func() {
		stopCtx, c := context.WithTimeout(t.Context(), 5*time.Second)
		defer c()
		relay.Stop(stopCtx)
	})

	require.Eventually(t, func() bool {
		var msg OutboxMessage
		db.First(&msg, 1)
		return msg.Status == StatusSent
	}, 10*time.Second, 100*time.Millisecond)

	sent := producer.sentMessages()
	require.Len(t, sent, 1)
	assert.Equal(t, "orders", sent[0].Topic)
	assert.Equal(t, []byte("key-1"), sent[0].Key)

	var msg OutboxMessage
	db.First(&msg, 1)
	assert.Equal(t, StatusSent, msg.Status)
	assert.NotNil(t, msg.SentAt)
}

func TestRelay_SendFailure(t *testing.T) {
	store, db := setupTestStore(t)
	producer := newMockPublisher()
	producer.sendErr = errors.New("connection refused")

	relay, err := NewRelay(store, producer,
		WithPollInterval(50*time.Millisecond),
		WithCleanupInterval(time.Hour),
	)
	require.NoError(t, err)

	ctx := t.Context()

	db.Create(&OutboxMessage{
		Topic: "events",
		Value: []byte("data"),
	})

	require.NoError(t, relay.Start(ctx))
	t.Cleanup(func() {
		stopCtx, c := context.WithTimeout(t.Context(), 5*time.Second)
		defer c()
		relay.Stop(stopCtx)
	})

	require.Eventually(t, func() bool {
		var msg OutboxMessage
		db.First(&msg, 1)
		return msg.Status == StatusFailed
	}, 10*time.Second, 100*time.Millisecond)

	var msg OutboxMessage
	db.First(&msg, 1)
	assert.Equal(t, StatusFailed, msg.Status)
	assert.Equal(t, "connection refused", msg.LastError)
	assert.Equal(t, 1, msg.RetryCount)
}

func TestRelay_MaxRetriesSkip(t *testing.T) {
	store, db := setupTestStore(t)
	producer := newMockPublisher()

	relay, err := NewRelay(store, producer,
		WithPollInterval(50*time.Millisecond),
		WithCleanupInterval(time.Hour),
		WithMaxRetries(2),
	)
	require.NoError(t, err)

	ctx := t.Context()

	db.Create(&OutboxMessage{
		Topic:      "events",
		Value:      []byte("data"),
		RetryCount: 2,
	})

	require.NoError(t, relay.Start(ctx))
	t.Cleanup(func() {
		stopCtx, c := context.WithTimeout(t.Context(), 5*time.Second)
		defer c()
		relay.Stop(stopCtx)
	})

	time.Sleep(200 * time.Millisecond)

	assert.Empty(t, producer.sentMessages())
}

// --- Options 测试 ---

func TestDefaultOptions(t *testing.T) {
	opts := defaultOptions()
	assert.Equal(t, time.Second, opts.pollInterval)
	assert.Equal(t, 100, opts.batchSize)
	assert.Equal(t, 3, opts.maxRetries)
	assert.Equal(t, 7*24*time.Hour, opts.cleanupAge)
	assert.Equal(t, time.Hour, opts.cleanupInterval)
	assert.Equal(t, 5*time.Minute, opts.staleTimeout)
	assert.Nil(t, opts.logger)
}

func TestApplyOptions(t *testing.T) {
	opts := applyOptions([]Option{
		WithPollInterval(2 * time.Second),
		WithBatchSize(50),
		WithMaxRetries(5),
		WithCleanupAge(24 * time.Hour),
		WithCleanupInterval(30 * time.Minute),
		WithStaleTimeout(10 * time.Minute),
	})

	assert.Equal(t, 2*time.Second, opts.pollInterval)
	assert.Equal(t, 50, opts.batchSize)
	assert.Equal(t, 5, opts.maxRetries)
	assert.Equal(t, 24*time.Hour, opts.cleanupAge)
	assert.Equal(t, 30*time.Minute, opts.cleanupInterval)
	assert.Equal(t, 10*time.Minute, opts.staleTimeout)
}

// --- Errors 测试 ---

func TestErrors(t *testing.T) {
	assert.True(t, errors.Is(ErrNilStore, ErrNilStore))
	assert.True(t, errors.Is(ErrNilProducer, ErrNilProducer))
	assert.True(t, errors.Is(ErrRelayAlreadyRunning, ErrRelayAlreadyRunning))
	assert.True(t, errors.Is(ErrRelayNotRunning, ErrRelayNotRunning))
	assert.True(t, errors.Is(ErrEmptyTopic, ErrEmptyTopic))
	assert.True(t, errors.Is(ErrEmptyValue, ErrEmptyValue))
	assert.True(t, errors.Is(ErrNilDB, ErrNilDB))
}

package websocket

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.ReadBufferSize != 1024 {
		t.Errorf("ReadBufferSize = %d, want 1024", cfg.ReadBufferSize)
	}
	if cfg.WriteBufferSize != 1024 {
		t.Errorf("WriteBufferSize = %d, want 1024", cfg.WriteBufferSize)
	}
	if cfg.MaxMessageSize != 512*1024 {
		t.Errorf("MaxMessageSize = %d, want %d", cfg.MaxMessageSize, 512*1024)
	}
	if cfg.WriteTimeout != 10*time.Second {
		t.Errorf("WriteTimeout = %v", cfg.WriteTimeout)
	}
	if cfg.PingInterval != 30*time.Second {
		t.Errorf("PingInterval = %v", cfg.PingInterval)
	}
	if !cfg.EnableCompression {
		t.Error("expected EnableCompression=true")
	}
}

func TestNewHub(t *testing.T) {
	handler := func(client Client, msg *Message) {}
	h := NewHub(handler)
	if h == nil {
		t.Fatal("NewHub returned nil")
	}
	if h.Count() != 0 {
		t.Errorf("Count = %d, want 0", h.Count())
	}
	if len(h.Clients()) != 0 {
		t.Errorf("Clients len = %d, want 0", len(h.Clients()))
	}
}

func TestNewHubNilHandler(t *testing.T) {
	h := NewHub(nil)
	if h == nil {
		t.Fatal("NewHub returned nil")
	}
}

func TestNewHubWithMiddleware(t *testing.T) {
	mw := func(next Handler) Handler {
		return func(client Client, msg *Message) { next(client, msg) }
	}
	handler := func(client Client, msg *Message) {}
	h := NewHub(handler, mw)
	if h == nil {
		t.Fatal("NewHub returned nil")
	}
}

func TestHubRunAndClose(t *testing.T) {
	h := NewHub(nil)
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() { done <- h.Run(ctx) }()

	time.Sleep(10 * time.Millisecond)
	cancel()

	if err := <-done; err != nil {
		t.Fatalf("Run error: %v", err)
	}
}

func TestHubSendToUnknownClient(t *testing.T) {
	h := NewHub(nil)
	err := h.Send("nonexistent", &Message{Type: TextMessage, Data: []byte("hi")})
	if err != ErrClientNotFound {
		t.Fatalf("expected ErrClientNotFound, got %v", err)
	}
}

func TestHubClientLookup(t *testing.T) {
	h := NewHub(nil)
	_, ok := h.Client("no-such-id")
	if ok {
		t.Fatal("expected false for nonexistent client")
	}
}

func TestMessageTypes(t *testing.T) {
	if TextMessage != 1 {
		t.Errorf("TextMessage = %d", TextMessage)
	}
	if BinaryMessage != 2 {
		t.Errorf("BinaryMessage = %d", BinaryMessage)
	}
	if CloseMessage != 8 {
		t.Errorf("CloseMessage = %d", CloseMessage)
	}
	if PingMessage != 9 {
		t.Errorf("PingMessage = %d", PingMessage)
	}
	if PongMessage != 10 {
		t.Errorf("PongMessage = %d", PongMessage)
	}
}

func TestErrors(t *testing.T) {
	errs := []error{ErrClientNotFound, ErrHubClosed, ErrConnectionClosed, ErrWriteTimeout, ErrMessageTooLarge, ErrInvalidMessage, ErrUpgradeFailed}
	for _, e := range errs {
		if e == nil || e.Error() == "" {
			t.Errorf("error should not be nil/empty: %v", e)
		}
	}
}

func TestHubCloseIdempotent(t *testing.T) {
	h := NewHub(nil)
	if err := h.Close(); err != nil {
		t.Fatalf("Close error: %v", err)
	}
	if err := h.Close(); err != nil {
		t.Fatalf("second Close error: %v", err)
	}
}

func TestDefaultConfigValues(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.ReadTimeout != 60*time.Second {
		t.Errorf("ReadTimeout = %v, want 60s", cfg.ReadTimeout)
	}
	if cfg.PongTimeout != 60*time.Second {
		t.Errorf("PongTimeout = %v, want 60s", cfg.PongTimeout)
	}
	if cfg.CheckOrigin != nil {
		t.Error("CheckOrigin should be nil by default")
	}
}

// --- mock client (thread-safe) ---

type mockClient struct {
	id       string
	ctx      context.Context
	metadata map[string]any
	mu       sync.Mutex
	sent     []*Message
	closed   bool
}

func newMockClient(id string) *mockClient {
	return &mockClient{id: id, ctx: context.Background(), metadata: make(map[string]any)}
}

func (c *mockClient) ID() string { return c.id }

func (c *mockClient) Send(msg *Message) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.sent = append(c.sent, msg)
	return nil
}

func (c *mockClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closed = true
	return nil
}

func (c *mockClient) Context() context.Context          { return c.ctx }
func (c *mockClient) SetContext(ctx context.Context)    { c.ctx = ctx }
func (c *mockClient) Metadata() map[string]any          { return c.metadata }
func (c *mockClient) SetMetadata(key string, value any) { c.metadata[key] = value }

func (c *mockClient) getSentCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.sent)
}

func (c *mockClient) getSent() []*Message {
	c.mu.Lock()
	defer c.mu.Unlock()
	cp := make([]*Message, len(c.sent))
	copy(cp, c.sent)
	return cp
}

func (c *mockClient) getClosed() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.closed
}

// --- Hub integration tests ---

func TestHubRegisterAndBroadcast(t *testing.T) {
	h := NewHub(nil)
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() { done <- h.Run(ctx) }()

	c1 := newMockClient("c1")
	c2 := newMockClient("c2")
	h.Register(c1)
	h.Register(c2)
	time.Sleep(20 * time.Millisecond)

	if h.Count() != 2 {
		t.Fatalf("Count = %d, want 2", h.Count())
	}

	h.Broadcast(&Message{Type: TextMessage, Data: []byte("hello")})
	time.Sleep(20 * time.Millisecond)

	if c1.getSentCount() == 0 {
		t.Error("c1 should have received broadcast")
	}
	if c2.getSentCount() == 0 {
		t.Error("c2 should have received broadcast")
	}

	cancel()
	<-done
}

func TestHubUnregister(t *testing.T) {
	h := NewHub(nil)
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() { done <- h.Run(ctx) }()

	c := newMockClient("c1")
	h.Register(c)
	time.Sleep(20 * time.Millisecond)

	h.Unregister(c)
	time.Sleep(20 * time.Millisecond)

	if h.Count() != 0 {
		t.Errorf("Count after unregister = %d, want 0", h.Count())
	}
	if !c.getClosed() {
		t.Error("client should be closed after unregister")
	}

	cancel()
	<-done
}

func TestHubBroadcastTo(t *testing.T) {
	h := NewHub(nil)
	c1 := newMockClient("c1")
	c2 := newMockClient("c2")
	h.Register(c1)
	h.Register(c2)

	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() { done <- h.Run(ctx) }()
	time.Sleep(20 * time.Millisecond)

	h.BroadcastTo([]string{"c1"}, &Message{Type: TextMessage, Data: []byte("targeted")})
	time.Sleep(20 * time.Millisecond)

	if c1.getSentCount() == 0 {
		t.Error("c1 should have received BroadcastTo message")
	}
	c2Targeted := false
	for _, m := range c2.getSent() {
		if string(m.Data) == "targeted" {
			c2Targeted = true
		}
	}
	if c2Targeted {
		t.Error("c2 should NOT have received targeted message")
	}

	cancel()
	<-done
}

func TestHubSendToKnownClient(t *testing.T) {
	h := NewHub(nil)
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() { done <- h.Run(ctx) }()

	c := newMockClient("c1")
	h.Register(c)
	time.Sleep(20 * time.Millisecond)

	if err := h.Send("c1", &Message{Type: TextMessage, Data: []byte("direct")}); err != nil {
		t.Fatalf("Send error: %v", err)
	}
	time.Sleep(10 * time.Millisecond)

	sent := c.getSent()
	if len(sent) == 0 || string(sent[len(sent)-1].Data) != "direct" {
		t.Error("c1 should have received direct message")
	}

	cancel()
	<-done
}

func TestHubCloseDisconnectsClients(t *testing.T) {
	h := NewHub(nil)
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() { done <- h.Run(ctx) }()

	c := newMockClient("c1")
	h.Register(c)
	time.Sleep(20 * time.Millisecond)

	cancel()
	<-done

	if !c.getClosed() {
		t.Error("client should be closed when hub closes")
	}
}

func TestHubClients(t *testing.T) {
	h := NewHub(nil)
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() { done <- h.Run(ctx) }()

	h.Register(newMockClient("a"))
	h.Register(newMockClient("b"))
	time.Sleep(20 * time.Millisecond)

	if len(h.Clients()) != 2 {
		t.Errorf("Clients() len = %d, want 2", len(h.Clients()))
	}
	if _, ok := h.Client("a"); !ok {
		t.Error("Client('a') should exist")
	}

	cancel()
	<-done
}

func TestHubHandleMessage(t *testing.T) {
	var handled bool
	handler := func(client Client, msg *Message) { handled = true }
	h := NewHub(handler).(*hub)
	h.HandleMessage(newMockClient("c1"), &Message{Type: TextMessage, Data: []byte("test")})
	if !handled {
		t.Error("handler should have been called")
	}
}

func TestHubHandleMessageNilHandler(t *testing.T) {
	h := NewHub(nil).(*hub)
	h.HandleMessage(newMockClient("c"), &Message{})
}

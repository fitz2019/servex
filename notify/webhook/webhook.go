// Package webhook 提供 Webhook 事件投递与接收能力.
package webhook

import (
	"context"
	"net/http"
	"time"
)

// Event 表示一个 webhook 事件.
type Event struct {
	ID        string
	Type      string
	Payload   []byte
	Timestamp time.Time
}

// Subscription 表示一个 webhook 订阅.
type Subscription struct {
	ID       string
	URL      string
	Secret   string
	Events   []string
	Metadata map[string]string
}

// Dispatcher 投递 webhook 事件.
type Dispatcher interface {
	Dispatch(ctx context.Context, sub *Subscription, event *Event) error
	Close() error
}

// Receiver 接收并验证 webhook 请求.
type Receiver interface {
	Handle(ctx context.Context, r *http.Request) (*Event, error)
}

// SubscriptionStore 管理 webhook 订阅.
type SubscriptionStore interface {
	Save(ctx context.Context, sub *Subscription) error
	Delete(ctx context.Context, id string) error
	ListByEvent(ctx context.Context, eventType string) ([]*Subscription, error)
	Get(ctx context.Context, id string) (*Subscription, error)
}

// Signer 签名和验签.
type Signer interface {
	Sign(payload []byte, secret string) string
	Verify(payload []byte, secret string, signature string) bool
}

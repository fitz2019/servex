package notify

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/google/uuid"

	"github.com/Tsukikage7/servex/messaging/jobqueue"
)

type Dispatcher struct {
	opts    dispatcherOptions
	senders map[Channel]Sender
	mu      sync.RWMutex
	closed  atomic.Bool
}

func NewDispatcher(opts ...Option) *Dispatcher {
	var o dispatcherOptions
	for _, opt := range opts {
		opt(&o)
	}
	return &Dispatcher{opts: o, senders: make(map[Channel]Sender)}
}

func (d *Dispatcher) Register(sender Sender) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.senders[sender.Channel()] = sender
}

func (d *Dispatcher) Send(ctx context.Context, msg *Message) (*Result, error) {
	if d.closed.Load() {
		return nil, ErrClosed
	}
	if msg != nil && msg.Channel == "" && d.opts.defaultChannel != "" {
		msg.Channel = d.opts.defaultChannel
	}
	if err := ValidateMessage(msg); err != nil {
		return nil, err
	}

	if msg.TemplateID != "" && d.opts.templateEngine != nil {
		rendered, err := d.opts.templateEngine.Render(msg.TemplateID, msg.TemplateData)
		if err != nil {
			return nil, err
		}
		msg.Body = rendered
	}

	d.mu.RLock()
	sender, ok := d.senders[msg.Channel]
	d.mu.RUnlock()

	if !ok {
		return nil, ErrNoSender
	}
	return sender.Send(ctx, msg)
}

func (d *Dispatcher) Broadcast(ctx context.Context, channels []Channel, msg *Message) []*Result {
	results := make([]*Result, 0, len(channels))
	for _, ch := range channels {
		clone := *msg
		clone.Channel = ch
		result, err := d.Send(ctx, &clone)
		if err != nil {
			results = append(results, &Result{Channel: ch, Error: err})
		} else {
			results = append(results, result)
		}
	}
	return results
}

// SendAsync 将消息序列化后投入 jobqueue 异步发送。
func (d *Dispatcher) SendAsync(ctx context.Context, msg *Message) error {
	if d.closed.Load() {
		return ErrClosed
	}
	if msg != nil && msg.Channel == "" && d.opts.defaultChannel != "" {
		msg.Channel = d.opts.defaultChannel
	}
	if err := ValidateMessage(msg); err != nil {
		return err
	}
	if d.opts.jobClient == nil {
		return fmt.Errorf("notification: jobqueue 未配置")
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("notification: 序列化消息失败: %w", err)
	}

	return d.opts.jobClient.Enqueue(ctx, &jobqueue.Job{
		ID:      uuid.New().String(),
		Queue:   "notifications",
		Type:    "notification." + string(msg.Channel),
		Payload: payload,
	})
}

func (d *Dispatcher) Close() error {
	if d.closed.Swap(true) {
		return nil
	}
	d.mu.RLock()
	defer d.mu.RUnlock()
	var firstErr error
	for _, sender := range d.senders {
		if err := sender.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// notification/push/sender.go
package push

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/Tsukikage7/servex/notify"
)

type Sender struct {
	provider Provider
	opts     senderOptions
	closed   atomic.Bool
}

func NewSender(provider Provider, opts ...Option) (*Sender, error) {
	if provider == nil {
		return nil, errors.New("notification/push: provider 不能为空")
	}
	var o senderOptions
	for _, opt := range opts {
		opt(&o)
	}
	return &Sender{provider: provider, opts: o}, nil
}

func (s *Sender) Channel() notify.Channel { return notify.ChannelPush }

func (s *Sender) Send(ctx context.Context, msg *notify.Message) (*notify.Result, error) {
	if msg == nil {
		return nil, notify.ErrNilMessage
	}
	if s.closed.Load() {
		return nil, notify.ErrClosed
	}

	payload := &Payload{Title: msg.Subject, Body: msg.Body}
	if msg.Metadata != nil {
		if b, ok := msg.Metadata["badge"]; ok {
			if n, err := strconv.Atoi(b); err == nil {
				payload.Badge = n
			}
		}
		payload.Sound = msg.Metadata["sound"]
	}
	if len(msg.TemplateData) > 0 {
		payload.Data = make(map[string]string, len(msg.TemplateData))
		for k, v := range msg.TemplateData {
			payload.Data[k] = fmt.Sprintf("%v", v)
		}
	}

	var lastID string
	var errs []string
	for _, token := range msg.To {
		id, err := s.provider.Send(ctx, token, payload)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", token, err))
			continue
		}
		lastID = id
	}
	if len(errs) > 0 {
		return nil, fmt.Errorf("notification/push: 部分发送失败: %s", strings.Join(errs, "; "))
	}
	return &notify.Result{MessageID: lastID, Channel: notify.ChannelPush}, nil
}

func (s *Sender) Close() error { s.closed.Store(true); return nil }

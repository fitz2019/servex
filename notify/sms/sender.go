// notification/sms/sender.go
package sms

import (
	"context"
	"errors"
	"fmt"
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
		return nil, errors.New("notification/sms: provider 不能为空")
	}
	var o senderOptions
	for _, opt := range opts {
		opt(&o)
	}
	return &Sender{provider: provider, opts: o}, nil
}

func (s *Sender) Channel() notify.Channel { return notify.ChannelSMS }

func (s *Sender) Send(ctx context.Context, msg *notify.Message) (*notify.Result, error) {
	if msg == nil {
		return nil, notify.ErrNilMessage
	}
	if s.closed.Load() {
		return nil, notify.ErrClosed
	}

	params := make(map[string]string, len(msg.TemplateData))
	for k, v := range msg.TemplateData {
		params[k] = fmt.Sprintf("%v", v)
	}

	var lastID string
	var errs []string
	for _, phone := range msg.To {
		id, err := s.provider.Send(ctx, &SendRequest{
			Phone: phone, Content: msg.Body, SignName: s.opts.signName,
			TemplateCode: msg.TemplateID, Params: params,
		})
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", phone, err))
			continue
		}
		lastID = id
	}
	if len(errs) > 0 {
		return nil, fmt.Errorf("notification/sms: 部分发送失败: %s", strings.Join(errs, "; "))
	}
	return &notify.Result{MessageID: lastID, Channel: notify.ChannelSMS}, nil
}

func (s *Sender) Close() error { s.closed.Store(true); return nil }

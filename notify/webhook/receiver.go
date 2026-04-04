// webhook/receiver.go
package webhook

import (
	"context"
	"io"
	"net/http"
	"time"
)

type receiver struct {
	opts receiverOptions
}

// NewReceiver 创建 webhook 接收器。
func NewReceiver(opts ...ReceiverOption) *receiver {
	o := receiverOptions{
		signer:          NewHMACSigner(),
		signatureHeader: "X-Webhook-Signature",
		eventTypeHeader: "X-Webhook-Event",
		eventIDHeader:   "X-Webhook-ID",
	}
	for _, opt := range opts {
		opt(&o)
	}
	return &receiver{opts: o}
}

func (rc *receiver) Handle(ctx context.Context, r *http.Request) (*Event, error) {
	if r.Body == nil {
		return nil, ErrEmptyBody
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	if len(body) == 0 {
		return nil, ErrEmptyBody
	}

	// 有 secret 时验签
	if rc.opts.secret != "" {
		sig := r.Header.Get(rc.opts.signatureHeader)
		if !rc.opts.signer.Verify(body, rc.opts.secret, sig) {
			return nil, ErrInvalidSignature
		}
	}

	return &Event{
		ID:        r.Header.Get(rc.opts.eventIDHeader),
		Type:      r.Header.Get(rc.opts.eventTypeHeader),
		Payload:   body,
		Timestamp: time.Now(),
	}, nil
}

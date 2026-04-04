// webhook/dispatcher.go
package webhook

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"
)

type dispatcher struct {
	opts dispatcherOptions
}

// NewDispatcher 创建 webhook 投递器。
func NewDispatcher(opts ...DispatcherOption) *dispatcher {
	o := dispatcherOptions{
		timeout:         10 * time.Second,
		signer:          NewHMACSigner(),
		signatureHeader: "X-Webhook-Signature",
		eventTypeHeader: "X-Webhook-Event",
		eventIDHeader:   "X-Webhook-ID",
	}
	for _, opt := range opts {
		opt(&o)
	}
	if o.httpClient == nil {
		o.httpClient = &http.Client{Timeout: o.timeout}
	}
	return &dispatcher{opts: o}
}

func (d *dispatcher) Dispatch(ctx context.Context, sub *Subscription, event *Event) error {
	if sub == nil {
		return ErrNilSubscription
	}
	if event == nil {
		return ErrNilEvent
	}
	if sub.URL == "" {
		return ErrEmptyURL
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, sub.URL, bytes.NewReader(event.Payload))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(d.opts.eventTypeHeader, event.Type)
	req.Header.Set(d.opts.eventIDHeader, event.ID)

	if sub.Secret != "" {
		sig := d.opts.signer.Sign(event.Payload, sub.Secret)
		req.Header.Set(d.opts.signatureHeader, sig)
	}

	resp, err := d.opts.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook: 投递失败，状态码 %d", resp.StatusCode)
	}
	return nil
}

func (d *dispatcher) Close() error { return nil }

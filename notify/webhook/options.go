// webhook/options.go
package webhook

import (
	"net/http"
	"time"
)

type dispatcherOptions struct {
	httpClient      *http.Client
	timeout         time.Duration
	signer          Signer
	signatureHeader string
	eventTypeHeader string
	eventIDHeader   string
}

type DispatcherOption func(*dispatcherOptions)

func WithHTTPClient(client *http.Client) DispatcherOption {
	return func(o *dispatcherOptions) { o.httpClient = client }
}

func WithTimeout(d time.Duration) DispatcherOption {
	return func(o *dispatcherOptions) { o.timeout = d }
}

func WithSigner(s Signer) DispatcherOption {
	return func(o *dispatcherOptions) { o.signer = s }
}

func WithSignatureHeader(header string) DispatcherOption {
	return func(o *dispatcherOptions) { o.signatureHeader = header }
}

type receiverOptions struct {
	signer          Signer
	secret          string
	signatureHeader string
	eventTypeHeader string
	eventIDHeader   string
}

type ReceiverOption func(*receiverOptions)

func WithReceiverSigner(s Signer) ReceiverOption {
	return func(o *receiverOptions) { o.signer = s }
}

func WithSecret(secret string) ReceiverOption {
	return func(o *receiverOptions) { o.secret = secret }
}

func WithReceiverSignatureHeader(header string) ReceiverOption {
	return func(o *receiverOptions) { o.signatureHeader = header }
}

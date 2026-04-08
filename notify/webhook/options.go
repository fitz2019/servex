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

// DispatcherOption 投递器配置选项.
type DispatcherOption func(*dispatcherOptions)

// WithHTTPClient 设置自定义 HTTP 客户端.
func WithHTTPClient(client *http.Client) DispatcherOption {
	return func(o *dispatcherOptions) { o.httpClient = client }
}

// WithTimeout 设置 HTTP 请求超时时间.
func WithTimeout(d time.Duration) DispatcherOption {
	return func(o *dispatcherOptions) { o.timeout = d }
}

// WithSigner 设置签名器.
func WithSigner(s Signer) DispatcherOption {
	return func(o *dispatcherOptions) { o.signer = s }
}

// WithSignatureHeader 设置签名请求头名称.
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

// ReceiverOption 接收器配置选项.
type ReceiverOption func(*receiverOptions)

// WithReceiverSigner 设置接收器签名器.
func WithReceiverSigner(s Signer) ReceiverOption {
	return func(o *receiverOptions) { o.signer = s }
}

// WithSecret 设置接收器验签密钥.
func WithSecret(secret string) ReceiverOption {
	return func(o *receiverOptions) { o.secret = secret }
}

// WithReceiverSignatureHeader 设置接收器签名请求头名称.
func WithReceiverSignatureHeader(header string) ReceiverOption {
	return func(o *receiverOptions) { o.signatureHeader = header }
}

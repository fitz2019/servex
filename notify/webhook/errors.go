package webhook

import "errors"

// ErrNilSubscription 订阅为空.
// ErrNilEvent 事件为空.
// ErrEmptyURL 订阅 URL 为空.
// ErrInvalidSignature 签名验证失败.
// ErrEmptyBody 请求体为空.
// ErrNotFound 未找到.
var (
	ErrNilSubscription  = errors.New("webhook: subscription 为空")
	ErrNilEvent         = errors.New("webhook: event 为空")
	ErrEmptyURL         = errors.New("webhook: subscription URL 为空")
	ErrInvalidSignature = errors.New("webhook: 签名验证失败")
	ErrEmptyBody        = errors.New("webhook: 请求体为空")
	ErrNotFound         = errors.New("webhook: 未找到")
)

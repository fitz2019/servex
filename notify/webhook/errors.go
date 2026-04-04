// webhook/errors.go
package webhook

import "errors"

var (
	ErrNilSubscription  = errors.New("webhook: subscription 为空")
	ErrNilEvent         = errors.New("webhook: event 为空")
	ErrEmptyURL         = errors.New("webhook: subscription URL 为空")
	ErrInvalidSignature = errors.New("webhook: 签名验证失败")
	ErrEmptyBody        = errors.New("webhook: 请求体为空")
	ErrNotFound         = errors.New("webhook: 未找到")
)

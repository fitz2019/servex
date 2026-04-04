// notification/sms/provider.go
package sms

import "context"

// Provider 短信服务商接口。
type Provider interface {
	Send(ctx context.Context, req *SendRequest) (messageID string, err error)
	Name() string
}

// SendRequest 短信发送请求。
type SendRequest struct {
	Phone        string
	Content      string
	SignName     string
	TemplateCode string
	Params       map[string]string
}

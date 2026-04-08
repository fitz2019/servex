// Package sms 提供短信发送能力，支持阿里云和腾讯云.
package sms

import "context"

// Provider 短信服务商接口.
type Provider interface {
	Send(ctx context.Context, req *SendRequest) (messageID string, err error)
	Name() string
}

// SendRequest 短信发送请求.
type SendRequest struct {
	Phone        string
	Content      string
	SignName     string
	TemplateCode string
	Params       map[string]string
}

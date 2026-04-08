// Package push 提供移动端推送通知能力，支持 FCM 和 APNs.
package push

import "context"

// Provider 推送服务提供者接口.
type Provider interface {
	Send(ctx context.Context, token string, payload *Payload) (messageID string, err error)
	Name() string
}

// Payload 推送消息载荷.
type Payload struct {
	Title string
	Body  string
	Data  map[string]string
	Badge int
	Sound string
}

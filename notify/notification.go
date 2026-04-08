// Package notify 提供多渠道消息通知能力，支持邮件、短信、Webhook 和推送.
package notify

import "context"

// Channel 通知渠道类型.
type Channel string

// ChannelEmail 邮件渠道.
// ChannelSMS 短信渠道.
// ChannelWebhook Webhook 渠道.
// ChannelPush 推送渠道.
const (
	ChannelEmail   Channel = "email"
	ChannelSMS     Channel = "sms"
	ChannelWebhook Channel = "webhook"
	ChannelPush    Channel = "push"
)

var validChannels = map[Channel]bool{
	ChannelEmail: true, ChannelSMS: true,
	ChannelWebhook: true, ChannelPush: true,
}

// Valid 判断渠道是否有效.
func (c Channel) Valid() bool { return validChannels[c] }

// Message 通知消息.
type Message struct {
	Channel      Channel
	To           []string
	Subject      string
	Body         string
	TemplateID   string
	TemplateData map[string]any
	Metadata     map[string]string
}

// Result 发送结果.
type Result struct {
	MessageID string
	Channel   Channel
	Error     error
}

// Sender 消息发送器接口.
type Sender interface {
	Send(ctx context.Context, msg *Message) (*Result, error)
	Channel() Channel
	Close() error
}

// TemplateEngine 模板渲染引擎接口.
type TemplateEngine interface {
	Render(templateID string, data map[string]any) (string, error)
}

// ValidateMessage 校验消息参数.
func ValidateMessage(msg *Message) error {
	if msg == nil {
		return ErrNilMessage
	}
	if msg.Channel == "" {
		return ErrEmptyChannel
	}
	if !msg.Channel.Valid() {
		return ErrInvalidChannel
	}
	if len(msg.To) == 0 {
		return ErrEmptyRecipients
	}
	return nil
}

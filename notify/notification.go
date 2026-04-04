package notify

import "context"

type Channel string

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

func (c Channel) Valid() bool { return validChannels[c] }

type Message struct {
	Channel      Channel
	To           []string
	Subject      string
	Body         string
	TemplateID   string
	TemplateData map[string]any
	Metadata     map[string]string
}

type Result struct {
	MessageID string
	Channel   Channel
	Error     error
}

type Sender interface {
	Send(ctx context.Context, msg *Message) (*Result, error)
	Channel() Channel
	Close() error
}

type TemplateEngine interface {
	Render(templateID string, data map[string]any) (string, error)
}

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

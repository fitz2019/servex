package notify

import (
	"errors"
	"testing"
)

func TestChannel_String(t *testing.T) {
	tests := []struct {
		ch   Channel
		want string
	}{
		{ChannelEmail, "email"},
		{ChannelSMS, "sms"},
		{ChannelWebhook, "webhook"},
		{ChannelPush, "push"},
	}
	for _, tt := range tests {
		if got := string(tt.ch); got != tt.want {
			t.Errorf("Channel = %q, want %q", got, tt.want)
		}
	}
}

func TestChannel_Valid(t *testing.T) {
	if !ChannelEmail.Valid() {
		t.Error("email should be valid")
	}
	if Channel("fax").Valid() {
		t.Error("fax should not be valid")
	}
}

func TestMessage_Validate(t *testing.T) {
	tests := []struct {
		name string
		msg  *Message
		err  error
	}{
		{"nil message", nil, ErrNilMessage},
		{"empty channel", &Message{}, ErrEmptyChannel},
		{"invalid channel", &Message{Channel: "fax", To: []string{"x"}}, ErrInvalidChannel},
		{"empty recipients", &Message{Channel: ChannelEmail}, ErrEmptyRecipients},
		{"valid", &Message{Channel: ChannelEmail, To: []string{"a@b.com"}, Body: "hi"}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMessage(tt.msg)
			if !errors.Is(err, tt.err) {
				t.Errorf("got %v, want %v", err, tt.err)
			}
		})
	}
}

func TestErrors(t *testing.T) {
	errs := []error{
		ErrNilMessage, ErrEmptyChannel, ErrInvalidChannel,
		ErrEmptyRecipients, ErrNoSender, ErrClosed,
		ErrTemplateNotFound, ErrTemplateRender,
	}
	for _, e := range errs {
		if e == nil {
			t.Error("sentinel error should not be nil")
		}
	}
}

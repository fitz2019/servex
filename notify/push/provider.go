// notification/push/provider.go
package push

import "context"

type Provider interface {
	Send(ctx context.Context, token string, payload *Payload) (messageID string, err error)
	Name() string
}

type Payload struct {
	Title string
	Body  string
	Data  map[string]string
	Badge int
	Sound string
}

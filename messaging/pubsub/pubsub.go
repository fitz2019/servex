// pubsub/pubsub.go
package pubsub

import "context"

// Message 是 Pub/Sub 传输的基本单元。
type Message struct {
	ID       string
	Topic    string
	Key      []byte
	Body     []byte
	Headers  map[string]string
	Metadata map[string]any
}

// Publisher 将消息发布到指定 topic。
type Publisher interface {
	Publish(ctx context.Context, topic string, msgs ...*Message) error
	Close() error
}

// Subscriber 订阅 topic 并通过 channel 接收消息。
type Subscriber interface {
	Subscribe(ctx context.Context, topic string) (<-chan *Message, error)
	Ack(ctx context.Context, msg *Message) error
	Nack(ctx context.Context, msg *Message) error
	Close() error
}

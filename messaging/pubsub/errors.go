package pubsub

import "errors"

var (
	// ErrClosed 表示发布/订阅组件已关闭.
	ErrClosed = errors.New("pubsub: 已关闭")
	// ErrNilMessage 表示消息参数为空.
	ErrNilMessage = errors.New("pubsub: 消息为空")
	// ErrEmptyTopic 表示 topic 为空.
	ErrEmptyTopic = errors.New("pubsub: topic 为空")
	// ErrNoMessages 表示没有要发布的消息.
	ErrNoMessages = errors.New("pubsub: 没有要发布的消息")
	// ErrAckFailed 表示消息确认失败.
	ErrAckFailed = errors.New("pubsub: 确认失败")
	// ErrNackFailed 表示消息拒绝失败.
	ErrNackFailed = errors.New("pubsub: 拒绝失败")
)

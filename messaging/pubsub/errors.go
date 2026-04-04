// pubsub/errors.go
package pubsub

import "errors"

var (
	ErrClosed     = errors.New("pubsub: 已关闭")
	ErrNilMessage = errors.New("pubsub: 消息为空")
	ErrEmptyTopic = errors.New("pubsub: topic 为空")
	ErrNoMessages = errors.New("pubsub: 没有要发布的消息")
	ErrAckFailed  = errors.New("pubsub: 确认失败")
	ErrNackFailed = errors.New("pubsub: 拒绝失败")
)

package outbox

import "errors"

// 预定义错误.
//
// 所有错误均可通过 errors.Is 进行判断:
//
//	if errors.Is(err, outbox.ErrRelayAlreadyRunning) {
//	    // 处理中继器已运行的情况
//	}
var (
	// ErrNilStore Store 为空.
	ErrNilStore = errors.New("outbox: Store 为空")

	// ErrNilProducer Producer 为空.
	ErrNilProducer = errors.New("outbox: Producer 为空")

	// ErrRelayAlreadyRunning 中继器已运行.
	ErrRelayAlreadyRunning = errors.New("outbox: 中继器已运行")

	// ErrRelayNotRunning 中继器未运行.
	ErrRelayNotRunning = errors.New("outbox: 中继器未运行")

	// ErrEmptyTopic 消息主题为空.
	ErrEmptyTopic = errors.New("outbox: 消息主题为空")

	// ErrEmptyValue 消息内容为空.
	ErrEmptyValue = errors.New("outbox: 消息内容为空")

	// ErrNilDB 数据库实例为空.
	ErrNilDB = errors.New("outbox: 数据库实例为空")
)

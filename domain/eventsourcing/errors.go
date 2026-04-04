package eventsourcing

import "errors"

// 预定义错误.
var (
	// ErrNilEventStore 事件存储为空.
	ErrNilEventStore = errors.New("eventsourcing: event store is nil")

	// ErrNilFactory 聚合工厂为空.
	ErrNilFactory = errors.New("eventsourcing: factory is nil")

	// ErrAggregateNotFound 聚合不存在.
	ErrAggregateNotFound = errors.New("eventsourcing: aggregate not found")

	// ErrConcurrencyConflict 并发冲突.
	ErrConcurrencyConflict = errors.New("eventsourcing: concurrency conflict")

	// ErrNoEvents 没有事件需要保存.
	ErrNoEvents = errors.New("eventsourcing: no events to save")
)

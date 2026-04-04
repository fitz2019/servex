package middleware

import (
	"context"
	"sync"

	"github.com/Tsukikage7/servex/ai"
)

// UsageTracker 线程安全的 token 用量累计追踪器.
type UsageTracker struct {
	mu    sync.Mutex
	total ai.Usage
}

// Middleware 返回用量追踪中间件.
func (t *UsageTracker) Middleware() Middleware {
	return func(next ai.ChatModel) ai.ChatModel {
		return Wrap(
			func(ctx context.Context, messages []ai.Message, opts ...ai.CallOption) (*ai.ChatResponse, error) {
				resp, err := next.Generate(ctx, messages, opts...)
				if err == nil && resp != nil {
					t.mu.Lock()
					t.total.Add(resp.Usage)
					t.mu.Unlock()
				}
				return resp, err
			},
			func(ctx context.Context, messages []ai.Message, opts ...ai.CallOption) (ai.StreamReader, error) {
				// 流式接口：包装 StreamReader，在流结束后累计用量
				reader, err := next.Stream(ctx, messages, opts...)
				if err != nil {
					return nil, err
				}
				return &trackingStreamReader{reader: reader, tracker: t}, nil
			},
		)
	}
}

// Total 返回累计总用量（线程安全）.
func (t *UsageTracker) Total() ai.Usage {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.total
}

// Reset 重置累计用量（线程安全）.
func (t *UsageTracker) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.total = ai.Usage{}
}

// trackingStreamReader 在流结束后记录用量的 StreamReader 包装器.
type trackingStreamReader struct {
	reader  ai.StreamReader
	tracker *UsageTracker
	done    bool
}

// Recv 读取下一个片段，流结束时记录用量.
func (r *trackingStreamReader) Recv() (ai.StreamChunk, error) {
	chunk, err := r.reader.Recv()
	if err != nil && !r.done {
		r.done = true
		if resp := r.reader.Response(); resp != nil {
			r.tracker.mu.Lock()
			r.tracker.total.Add(resp.Usage)
			r.tracker.mu.Unlock()
		}
	}
	return chunk, err
}

// Response 获取完整响应.
func (r *trackingStreamReader) Response() *ai.ChatResponse {
	return r.reader.Response()
}

// Close 关闭流.
func (r *trackingStreamReader) Close() error {
	return r.reader.Close()
}

// 编译期接口断言.
var _ ai.StreamReader = (*trackingStreamReader)(nil)

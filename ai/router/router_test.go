package router_test

import (
	"context"
	"io"
	"testing"

	"github.com/Tsukikage7/servex/ai"
	"github.com/Tsukikage7/servex/ai/router"
)

// mockModel 记录调用时使用的 model 名称.
type mockModel struct {
	name string
}

func (m *mockModel) Generate(_ context.Context, _ []ai.Message, opts ...ai.CallOption) (*ai.ChatResponse, error) {
	return &ai.ChatResponse{Message: ai.AssistantMessage(m.name), ModelID: m.name}, nil
}

func (m *mockModel) Stream(_ context.Context, _ []ai.Message, opts ...ai.CallOption) (ai.StreamReader, error) {
	// 返回一个只含单条消息的 mock StreamReader
	return &mockStream{content: m.name}, nil
}

// mockStream 单条消息的流式读取器.
type mockStream struct {
	content string
	done    bool
}

func (s *mockStream) Recv() (ai.StreamChunk, error) {
	if s.done {
		return ai.StreamChunk{}, io.EOF
	}
	s.done = true
	return ai.StreamChunk{Delta: s.content, FinishReason: "stop"}, nil
}

func (s *mockStream) Response() *ai.ChatResponse { return nil }
func (s *mockStream) Close() error               { return nil }

func TestRouter_RoutesByModel(t *testing.T) {
	dashscope := &mockModel{name: "dashscope"}
	openai := &mockModel{name: "openai"}
	fallback := &mockModel{name: "fallback"}

	r := router.New(fallback,
		router.Route{Models: []string{"qwen-plus", "qwen-max"}, Model: dashscope},
		router.Route{Models: []string{"gpt-4o"}, Model: openai},
	)

	resp, err := r.Generate(t.Context(), nil, ai.WithModel("qwen-plus"))
	if err != nil {
		t.Fatalf("Generate 失败: %v", err)
	}
	if resp.ModelID != "dashscope" {
		t.Errorf("期望路由到 dashscope，得到 %q", resp.ModelID)
	}

	resp, err = r.Generate(t.Context(), nil, ai.WithModel("gpt-4o"))
	if err != nil {
		t.Fatalf("Generate 失败: %v", err)
	}
	if resp.ModelID != "openai" {
		t.Errorf("期望路由到 openai，得到 %q", resp.ModelID)
	}
}

func TestRouter_FallbackWhenNoMatch(t *testing.T) {
	fallback := &mockModel{name: "fallback"}
	r := router.New(fallback,
		router.Route{Models: []string{"known-model"}, Model: &mockModel{name: "other"}},
	)

	resp, err := r.Generate(t.Context(), nil, ai.WithModel("unknown-model"))
	if err != nil {
		t.Fatalf("Generate 失败: %v", err)
	}
	if resp.ModelID != "fallback" {
		t.Errorf("期望 fallback，得到 %q", resp.ModelID)
	}
}

func TestRouter_FallbackWhenNoModel(t *testing.T) {
	fallback := &mockModel{name: "fallback"}
	r := router.New(fallback,
		router.Route{Models: []string{"gpt-4o"}, Model: &mockModel{name: "other"}},
	)

	// 不传 WithModel，model 为空字符串
	resp, err := r.Generate(t.Context(), nil)
	if err != nil {
		t.Fatalf("Generate 失败: %v", err)
	}
	if resp.ModelID != "fallback" {
		t.Errorf("期望 fallback，得到 %q", resp.ModelID)
	}
}

func TestRouter_FirstMatchWins(t *testing.T) {
	first := &mockModel{name: "first"}
	second := &mockModel{name: "second"}
	fallback := &mockModel{name: "fallback"}

	r := router.New(fallback,
		router.Route{Models: []string{"shared-model"}, Model: first},
		router.Route{Models: []string{"shared-model"}, Model: second},
	)

	resp, err := r.Generate(t.Context(), nil, ai.WithModel("shared-model"))
	if err != nil {
		t.Fatalf("Generate 失败: %v", err)
	}
	if resp.ModelID != "first" {
		t.Errorf("期望第一个命中的路由（first），得到 %q", resp.ModelID)
	}
}

func TestRouter_Stream(t *testing.T) {
	target := &mockModel{name: "stream-target"}
	fallback := &mockModel{name: "fallback"}

	r := router.New(fallback,
		router.Route{Models: []string{"stream-model"}, Model: target},
	)

	stream, err := r.Stream(t.Context(), nil, ai.WithModel("stream-model"))
	if err != nil {
		t.Fatalf("Stream 失败: %v", err)
	}
	defer stream.Close()

	chunk, err := stream.Recv()
	if err != nil {
		t.Fatalf("Recv 失败: %v", err)
	}
	if chunk.Delta != "stream-target" {
		t.Errorf("期望流内容 'stream-target'，得到 %q", chunk.Delta)
	}

	// 无命中时用 fallback
	stream2, err := r.Stream(t.Context(), nil, ai.WithModel("other-model"))
	if err != nil {
		t.Fatalf("Stream 失败: %v", err)
	}
	defer stream2.Close()

	chunk2, err := stream2.Recv()
	if err != nil {
		t.Fatalf("Recv 失败: %v", err)
	}
	if chunk2.Delta != "fallback" {
		t.Errorf("期望流内容 'fallback'，得到 %q", chunk2.Delta)
	}
}

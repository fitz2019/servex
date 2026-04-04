package conversation_test

import (
	"context"
	"io"
	"testing"

	"github.com/Tsukikage7/servex/ai"
	"github.com/Tsukikage7/servex/ai/conversation"
)

// mockModel 模拟 ChatModel.
type mockModel struct {
	responses []string
	idx       int
}

func (m *mockModel) Generate(ctx context.Context, messages []ai.Message, opts ...ai.CallOption) (*ai.ChatResponse, error) {
	if m.idx >= len(m.responses) {
		return &ai.ChatResponse{Message: ai.AssistantMessage("default")}, nil
	}
	resp := m.responses[m.idx]
	m.idx++
	return &ai.ChatResponse{Message: ai.AssistantMessage(resp)}, nil
}

func (m *mockModel) Stream(ctx context.Context, messages []ai.Message, opts ...ai.CallOption) (ai.StreamReader, error) {
	return &mockReader{content: "stream"}, nil
}

// mockReader 模拟 StreamReader.
type mockReader struct {
	content string
	sent    bool
}

func (r *mockReader) Recv() (ai.StreamChunk, error) {
	if r.sent {
		return ai.StreamChunk{}, io.EOF
	}
	r.sent = true
	return ai.StreamChunk{Delta: r.content, FinishReason: "stop"}, nil
}

func (r *mockReader) Response() *ai.ChatResponse {
	return &ai.ChatResponse{Message: ai.AssistantMessage(r.content)}
}

func (r *mockReader) Close() error { return nil }

func TestConversation_Chat(t *testing.T) {
	model := &mockModel{responses: []string{"你好！", "我很好，谢谢！"}}
	conv := conversation.New(model)

	resp1, err := conv.Chat(t.Context(), "你好")
	if err != nil {
		t.Fatalf("第一次 Chat 失败: %v", err)
	}
	if resp1.Message.Content != "你好！" {
		t.Errorf("期望 '你好！'，得到 %q", resp1.Message.Content)
	}

	resp2, err := conv.Chat(t.Context(), "你好吗？")
	if err != nil {
		t.Fatalf("第二次 Chat 失败: %v", err)
	}
	if resp2.Message.Content != "我很好，谢谢！" {
		t.Errorf("期望 '我很好，谢谢！'，得到 %q", resp2.Message.Content)
	}

	// 验证历史保留
	history := conv.History()
	if len(history) != 4 {
		t.Errorf("期望 4 条历史消息（2 用户 + 2 助手），得到 %d", len(history))
	}
}

func TestConversation_WithSystemPrompt(t *testing.T) {
	var capturedMessages []ai.Message
	model := &mockCaptureModel{capture: &capturedMessages}
	conv := conversation.New(model, conversation.WithSystemPrompt("你是一个助手"))

	_, _ = conv.Chat(t.Context(), "hi")

	if len(capturedMessages) == 0 {
		t.Fatal("未捕获消息")
	}
	if capturedMessages[0].Role != ai.RoleSystem {
		t.Errorf("期望第一条消息为 system，得到 %s", capturedMessages[0].Role)
	}
	if capturedMessages[0].Content != "你是一个助手" {
		t.Errorf("期望系统提示 '你是一个助手'，得到 %q", capturedMessages[0].Content)
	}
}

func TestConversation_Reset(t *testing.T) {
	model := &mockModel{responses: []string{"ok"}}
	conv := conversation.New(model)

	_, _ = conv.Chat(t.Context(), "hi")
	if len(conv.History()) == 0 {
		t.Fatal("期望有历史消息")
	}

	conv.Reset()
	if len(conv.History()) != 0 {
		t.Errorf("Reset 后期望历史为空，得到 %d 条", len(conv.History()))
	}
}

// mockCaptureModel 捕获发送消息的模型.
type mockCaptureModel struct {
	capture *[]ai.Message
}

func (m *mockCaptureModel) Generate(ctx context.Context, messages []ai.Message, opts ...ai.CallOption) (*ai.ChatResponse, error) {
	*m.capture = append(*m.capture, messages...)
	return &ai.ChatResponse{Message: ai.AssistantMessage("ok")}, nil
}

func (m *mockCaptureModel) Stream(ctx context.Context, messages []ai.Message, opts ...ai.CallOption) (ai.StreamReader, error) {
	return &mockReader{content: "ok"}, nil
}

func TestBufferMemory(t *testing.T) {
	mem := conversation.NewBufferMemory()
	mem.Add(ai.UserMessage("hi"))
	mem.Add(ai.AssistantMessage("hello"))

	msgs := mem.Messages()
	if len(msgs) != 2 {
		t.Errorf("期望 2 条消息，得到 %d", len(msgs))
	}

	mem.Clear()
	if len(mem.Messages()) != 0 {
		t.Error("Clear 后期望为空")
	}
}

func TestWindowMemory_Trim(t *testing.T) {
	// 最大 2 轮 = 4 条消息
	mem := conversation.NewWindowMemory(2)

	for range 3 {
		mem.Add(ai.UserMessage("user"))
		mem.Add(ai.AssistantMessage("assistant"))
	}

	msgs := mem.Messages()
	// 只保留最近 2 轮（4 条）
	if len(msgs) != 4 {
		t.Errorf("期望 4 条消息（2 轮），得到 %d", len(msgs))
	}
}

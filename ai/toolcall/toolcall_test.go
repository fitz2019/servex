package toolcall_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/Tsukikage7/servex/ai"
	"github.com/Tsukikage7/servex/ai/toolcall"
)

// mockToolModel 模拟支持工具调用的模型.
type mockToolModel struct {
	rounds []ai.ChatResponse // 每轮的响应
	idx    int
}

func (m *mockToolModel) Generate(ctx context.Context, messages []ai.Message, opts ...ai.CallOption) (*ai.ChatResponse, error) {
	if m.idx >= len(m.rounds) {
		return &ai.ChatResponse{Message: ai.AssistantMessage("完成")}, nil
	}
	resp := m.rounds[m.idx]
	m.idx++
	return &resp, nil
}

func (m *mockToolModel) Stream(ctx context.Context, messages []ai.Message, opts ...ai.CallOption) (ai.StreamReader, error) {
	return nil, nil
}

func TestRegistry_Register(t *testing.T) {
	reg := toolcall.NewRegistry()

	tool := ai.Tool{
		Function: ai.FunctionDef{
			Name:        "add",
			Description: "两数相加",
			Parameters:  json.RawMessage(`{"type":"object","properties":{"a":{"type":"number"},"b":{"type":"number"}}}`),
		},
	}
	reg.Register(tool, func(ctx context.Context, args string) (string, error) {
		var params struct{ A, B float64 }
		json.Unmarshal([]byte(args), &params)
		result, _ := json.Marshal(map[string]float64{"result": params.A + params.B})
		return string(result), nil
	})

	tools := reg.Tools()
	if len(tools) != 1 {
		t.Errorf("期望 1 个工具，得到 %d", len(tools))
	}
	if tools[0].Function.Name != "add" {
		t.Errorf("期望工具名 'add'，得到 %q", tools[0].Function.Name)
	}
}

func TestExecutor_SingleToolCall(t *testing.T) {
	toolCall := ai.ToolCall{ID: "call_1"}
	toolCall.Function.Name = "get_time"
	toolCall.Function.Arguments = `{}`

	model := &mockToolModel{
		rounds: []ai.ChatResponse{
			// 第一轮：请求工具调用
			{
				Message:      ai.Message{Role: ai.RoleAssistant, ToolCalls: []ai.ToolCall{toolCall}},
				FinishReason: "tool_calls",
			},
			// 第二轮：最终回复
			{
				Message:      ai.AssistantMessage("当前时间是 12:00"),
				FinishReason: "stop",
			},
		},
	}

	reg := toolcall.NewRegistry()
	reg.Register(
		ai.Tool{Function: ai.FunctionDef{Name: "get_time"}},
		func(ctx context.Context, args string) (string, error) {
			return `{"time":"12:00"}`, nil
		},
	)

	executor := toolcall.NewExecutor(model, reg)
	result, err := executor.Run(t.Context(), []ai.Message{ai.UserMessage("现在几点？")})
	if err != nil {
		t.Fatalf("Run 失败: %v", err)
	}

	if result.Response.Message.Content != "当前时间是 12:00" {
		t.Errorf("期望最终回复 '当前时间是 12:00'，得到 %q", result.Response.Message.Content)
	}
	// Rounds 表示包含工具调用的轮次数（最终回复轮不计）
	if result.Rounds != 1 {
		t.Errorf("期望 1 个工具调用轮，得到 %d", result.Rounds)
	}

	// 验证历史包含工具调用和结果
	hasToolResult := false
	for _, msg := range result.Messages {
		if msg.Role == ai.RoleTool {
			hasToolResult = true
			break
		}
	}
	if !hasToolResult {
		t.Error("期望历史中包含工具调用结果消息")
	}
}

func TestExecutor_OnStepCallback(t *testing.T) {
	toolCall := ai.ToolCall{ID: "call_1"}
	toolCall.Function.Name = "get_time"
	toolCall.Function.Arguments = `{}`

	model := &mockToolModel{
		rounds: []ai.ChatResponse{
			// 第一轮：请求工具调用
			{
				Message:      ai.Message{Role: ai.RoleAssistant, ToolCalls: []ai.ToolCall{toolCall}},
				FinishReason: "tool_calls",
			},
			// 第二轮：最终回复
			{
				Message:      ai.AssistantMessage("完成"),
				FinishReason: "stop",
			},
		},
	}

	reg := toolcall.NewRegistry()
	reg.Register(
		ai.Tool{Function: ai.FunctionDef{Name: "get_time"}},
		func(ctx context.Context, args string) (string, error) {
			return `{"time":"12:00"}`, nil
		},
	)

	var events []toolcall.StepEvent
	executor := toolcall.NewExecutor(model, reg, toolcall.WithOnStep(func(ctx context.Context, event toolcall.StepEvent) {
		events = append(events, event)
	}))

	_, err := executor.Run(t.Context(), []ai.Message{ai.UserMessage("现在几点？")})
	if err != nil {
		t.Fatalf("Run 失败: %v", err)
	}

	if len(events) != 2 {
		t.Fatalf("期望触发 2 次回调，得到 %d", len(events))
	}

	// 第一次：工具调用轮
	e0 := events[0]
	if e0.Round != 0 {
		t.Errorf("第一次回调 Round 期望 0，得到 %d", e0.Round)
	}
	if e0.IsFinal {
		t.Error("第一次回调 IsFinal 期望 false")
	}
	if len(e0.ToolResults) != 1 {
		t.Errorf("第一次回调 ToolResults 期望 1 个，得到 %d", len(e0.ToolResults))
	}

	// 第二次：最终轮
	e1 := events[1]
	if e1.Round != 1 {
		t.Errorf("第二次回调 Round 期望 1，得到 %d", e1.Round)
	}
	if !e1.IsFinal {
		t.Error("第二次回调 IsFinal 期望 true")
	}
	if e1.ToolResults != nil {
		t.Errorf("第二次回调 ToolResults 期望 nil，得到 %v", e1.ToolResults)
	}
}

func TestExecutor_MaxRoundsExceeded(t *testing.T) {
	// 模型总是返回工具调用
	toolCall := ai.ToolCall{ID: "call_1"}
	toolCall.Function.Name = "infinite"
	toolCall.Function.Arguments = `{}`

	model := &mockToolModel{
		rounds: func() []ai.ChatResponse {
			rounds := make([]ai.ChatResponse, 20)
			for i := range rounds {
				rounds[i] = ai.ChatResponse{
					Message:      ai.Message{Role: ai.RoleAssistant, ToolCalls: []ai.ToolCall{toolCall}},
					FinishReason: "tool_calls",
				}
			}
			return rounds
		}(),
	}

	reg := toolcall.NewRegistry()
	reg.Register(
		ai.Tool{Function: ai.FunctionDef{Name: "infinite"}},
		func(ctx context.Context, args string) (string, error) {
			return `{}`, nil
		},
	)

	executor := toolcall.NewExecutor(model, reg, toolcall.WithMaxRounds(3))
	_, err := executor.Run(t.Context(), []ai.Message{ai.UserMessage("测试")})
	if err == nil {
		t.Fatal("期望超出轮次错误，得到 nil")
	}
}

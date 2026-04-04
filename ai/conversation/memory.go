// Package conversation 提供 AI 对话会话管理功能.
package conversation

import "github.com/Tsukikage7/servex/ai"

// Memory 记忆策略接口.
type Memory interface {
	// Add 添加一条消息到记忆.
	Add(msg ai.Message)
	// Messages 获取当前记忆中的所有消息.
	Messages() []ai.Message
	// Clear 清空记忆.
	Clear()
}

// BufferMemory 完整缓冲记忆，保留所有历史消息.
type BufferMemory struct {
	messages []ai.Message
}

// NewBufferMemory 创建完整缓冲记忆.
func NewBufferMemory() *BufferMemory {
	return &BufferMemory{}
}

// Add 添加消息.
func (m *BufferMemory) Add(msg ai.Message) {
	m.messages = append(m.messages, msg)
}

// Messages 获取所有消息.
func (m *BufferMemory) Messages() []ai.Message {
	result := make([]ai.Message, len(m.messages))
	copy(result, m.messages)
	return result
}

// Clear 清空记忆.
func (m *BufferMemory) Clear() {
	m.messages = nil
}

// 编译期接口断言.
var _ Memory = (*BufferMemory)(nil)

// WindowMemory 滑动窗口记忆，只保留最近 N 轮（每轮包含用户和助手各一条）对话.
type WindowMemory struct {
	maxRounds int
	messages  []ai.Message
}

// NewWindowMemory 创建滑动窗口记忆.
// maxRounds 为最大保留轮数（每轮 = 用户消息 + 助手消息）.
func NewWindowMemory(maxRounds int) *WindowMemory {
	if maxRounds < 1 {
		maxRounds = 1
	}
	return &WindowMemory{maxRounds: maxRounds}
}

// Add 添加消息，超出窗口时丢弃最旧的轮次.
func (m *WindowMemory) Add(msg ai.Message) {
	m.messages = append(m.messages, msg)
	m.trim()
}

// Messages 获取窗口内的消息.
func (m *WindowMemory) Messages() []ai.Message {
	result := make([]ai.Message, len(m.messages))
	copy(result, m.messages)
	return result
}

// Clear 清空记忆.
func (m *WindowMemory) Clear() {
	m.messages = nil
}

// trim 裁剪超出窗口的旧消息.
// 策略：每轮由 user+assistant 两条消息组成，最多保留 maxRounds*2 条.
func (m *WindowMemory) trim() {
	maxMessages := m.maxRounds * 2
	if len(m.messages) <= maxMessages {
		return
	}
	// 从最旧的 user 消息开始丢弃（保持 user/assistant 成对）
	excess := len(m.messages) - maxMessages
	m.messages = m.messages[excess:]
}

// 编译期接口断言.
var _ Memory = (*WindowMemory)(nil)

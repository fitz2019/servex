// Package prompt 提供基于 text/template 的 AI 消息模板引擎.
package prompt

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/Tsukikage7/servex/ai"
)

// Template AI 消息模板.
// 使用 Go text/template 语法，渲染后直接返回 ai.Message.
type Template struct {
	role ai.Role
	tmpl *template.Template
}

// New 创建消息模板.
// role 指定消息角色，text 为 Go text/template 格式的模板文本.
func New(role ai.Role, text string) (*Template, error) {
	tmpl, err := template.New("").Parse(text)
	if err != nil {
		return nil, fmt.Errorf("prompt: 解析模板失败: %w", err)
	}
	return &Template{role: role, tmpl: tmpl}, nil
}

// MustNew 创建消息模板，失败时 panic.
func MustNew(role ai.Role, text string) *Template {
	t, err := New(role, text)
	if err != nil {
		panic(err)
	}
	return t
}

// Render 使用 data 渲染模板，返回 ai.Message.
// data 可以是 struct、map 或任意 text/template 支持的数据类型.
func (t *Template) Render(data any) (ai.Message, error) {
	var buf bytes.Buffer
	if err := t.tmpl.Execute(&buf, data); err != nil {
		return ai.Message{}, fmt.Errorf("prompt: 渲染模板失败: %w", err)
	}
	return ai.Message{Role: t.role, Content: buf.String()}, nil
}

// MustRender 使用 data 渲染模板，失败时 panic.
func (t *Template) MustRender(data any) ai.Message {
	msg, err := t.Render(data)
	if err != nil {
		panic(err)
	}
	return msg
}

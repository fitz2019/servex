package prompt_test

import (
	"strings"
	"testing"

	"github.com/Tsukikage7/servex/ai"
	"github.com/Tsukikage7/servex/ai/prompt"
)

func TestTemplate_Render(t *testing.T) {
	tmpl, err := prompt.New(ai.RoleUser, "你好，{{.Name}}！今天是 {{.Day}}。")
	if err != nil {
		t.Fatalf("创建模板失败: %v", err)
	}

	data := struct {
		Name string
		Day  string
	}{"Alice", "星期一"}

	msg, err := tmpl.Render(data)
	if err != nil {
		t.Fatalf("渲染失败: %v", err)
	}

	if msg.Role != ai.RoleUser {
		t.Errorf("期望 Role=user，得到 %s", msg.Role)
	}
	if !strings.Contains(msg.Content, "Alice") {
		t.Errorf("期望内容包含 'Alice'，得到 %q", msg.Content)
	}
	if !strings.Contains(msg.Content, "星期一") {
		t.Errorf("期望内容包含 '星期一'，得到 %q", msg.Content)
	}
}

func TestTemplate_InvalidSyntax(t *testing.T) {
	_, err := prompt.New(ai.RoleSystem, "{{.Unclosed")
	if err == nil {
		t.Fatal("期望无效模板语法报错，得到 nil")
	}
}

func TestMustNew_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("期望 panic，但没有发生")
		}
	}()
	prompt.MustNew(ai.RoleSystem, "{{invalid")
}

func TestTemplate_WithMap(t *testing.T) {
	tmpl := prompt.MustNew(ai.RoleSystem, "语言: {{.lang}}")
	msg, err := tmpl.Render(map[string]string{"lang": "Go"})
	if err != nil {
		t.Fatalf("渲染失败: %v", err)
	}
	if msg.Content != "语言: Go" {
		t.Errorf("期望 '语言: Go'，得到 %q", msg.Content)
	}
}

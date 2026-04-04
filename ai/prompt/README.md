# ai/prompt

`ai/prompt` 包提供基于 Go `text/template` 的 AI 消息模板引擎，将模板渲染为 `ai.Message`。

## 功能特性

- 支持 Go `text/template` 全部语法（条件、循环、管道等）
- `Render(data)` 接受 struct、map 或任意类型
- `MustNew` / `MustRender` 变体在 panic 场景下简化代码

## 安装

```bash
go get github.com/Tsukikage7/servex/ai
```

## API

```go
func New(role ai.Role, text string) (*Template, error)
func MustNew(role ai.Role, text string) *Template

func (t *Template) Render(data any) (ai.Message, error)
func (t *Template) MustRender(data any) ai.Message
```

## 使用示例

```go
// 系统提示词模板
systemTmpl := prompt.MustNew(ai.RoleSystem,
    "你是一个专业的 {{.Language}} 工程师，擅长 {{.Domain}}。",
)

// 用户消息模板
userTmpl := prompt.MustNew(ai.RoleUser,
    `请审查以下代码并指出问题：
{{range .Files}}
文件：{{.Name}}
\`\`\`{{$.Language}}
{{.Content}}
\`\`\`
{{end}}`,
)

// 渲染
sysMsg := systemTmpl.MustRender(map[string]string{
    "Language": "Go",
    "Domain":   "微服务架构",
})

userMsg, err := userTmpl.Render(struct {
    Language string
    Files    []struct{ Name, Content string }
}{
    Language: "go",
    Files: []struct{ Name, Content string }{
        {"main.go", "package main\n..."},
    },
})

resp, _ := client.Generate(ctx, []ai.Message{sysMsg, userMsg})
```

## 许可证

详见项目根目录 LICENSE 文件。

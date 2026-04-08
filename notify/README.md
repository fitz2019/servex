# notify

## 导入路径

```go
import "github.com/Tsukikage7/servex/notify"
```

## 简介

`notify` 提供统一的通知发送抽象层，支持邮件、短信、Webhook、推送等多种渠道。核心类型包括 `Sender` 接口、`Message` 消息结构、`Dispatcher` 多渠道分发器和 `TemplateEngine` 模板引擎接口。

## 核心类型

| 类型 / 函数 | 说明 |
|---|---|
| `Channel` | 通知渠道常量（`ChannelEmail/SMS/Webhook/Push`） |
| `Message` | 消息结构（Channel/To/Subject/Body/TemplateID/Data/Metadata） |
| `Result` | 发送结果（MessageID/Channel/Success/Error/SentAt） |
| `Sender` | 发送者接口（`Send(ctx, msg) Result`） |
| `TemplateEngine` | 模板引擎接口（`Render(id, data) (subject, body, error)`） |
| `Dispatcher` | 多渠道分发器 |
| `NewDispatcher()` | 创建分发器 |
| `Dispatcher.Register(channel, sender)` | 注册渠道发送者 |
| `Dispatcher.Send(ctx, msg)` | 分发消息到对应渠道 |
| `ValidateMessage(msg)` | 校验消息合法性 |

## 示例

```go
package main

import (
    "context"
    "fmt"

    "github.com/Tsukikage7/servex/notify"
)

// MockSender 模拟发送者
type MockSender struct{}

func (s *MockSender) Send(ctx context.Context, msg notify.Message) notify.Result {
    fmt.Printf("[%s] 发送至 %s: %s\n", msg.Channel, msg.To, msg.Body)
    return notify.Result{
        Channel: msg.Channel,
        Success: true,
    }
}

func main() {
    dispatcher := notify.NewDispatcher()
    dispatcher.Register(notify.ChannelEmail, &MockSender{})
    dispatcher.Register(notify.ChannelSMS, &MockSender{})

    ctx := context.Background()

    // 发送邮件通知
    result := dispatcher.Send(ctx, notify.Message{
        Channel: notify.ChannelEmail,
        To:      "user@example.com",
        Subject: "订单确认",
        Body:    "您的订单 #12345 已确认。",
    })
    fmt.Println("邮件发送成功:", result.Success)

    // 发送短信通知
    result = dispatcher.Send(ctx, notify.Message{
        Channel: notify.ChannelSMS,
        To:      "+8613800138000",
        Body:    "您的验证码是 987654，5 分钟内有效。",
    })
    fmt.Println("短信发送成功:", result.Success)
}
```

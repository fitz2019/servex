package nwebhook

import "encoding/json"

// Formatter 消息格式化函数类型.
type Formatter func(subject, body string) []byte

func getFormatter(format string) Formatter {
	switch format {
	case "slack":
		return formatSlack
	case "dingtalk":
		return formatDingTalk
	case "lark":
		return formatLark
	default:
		return formatCustom
	}
}

func formatSlack(subject, body string) []byte {
	data, _ := json.Marshal(map[string]any{"text": subject + "\n" + body})
	return data
}

func formatDingTalk(subject, body string) []byte {
	data, _ := json.Marshal(map[string]any{
		"msgtype": "text", "text": map[string]string{"content": subject + "\n" + body},
	})
	return data
}

func formatLark(subject, body string) []byte {
	data, _ := json.Marshal(map[string]any{
		"msg_type": "text", "content": map[string]string{"text": subject + "\n" + body},
	})
	return data
}

func formatCustom(_, body string) []byte { return []byte(body) }

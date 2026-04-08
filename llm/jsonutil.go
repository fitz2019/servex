package llm

import (
	"strconv"
	"strings"
)

// ParseRetryAfter 解析 Retry-After 响应头.
func ParseRetryAfter(header string) int {
	v, err := strconv.Atoi(header)
	if err != nil {
		return 0
	}
	return v
}

// ExtractJSON 从可能包含 markdown 代码块的字符串中提取 JSON 内容.
func ExtractJSON(s string) string {
	s = strings.TrimSpace(s)
	// 处理 ```json ... ``` 代码块.
	if idx := strings.Index(s, "```json"); idx >= 0 {
		s = s[idx+7:]
		if end := strings.Index(s, "```"); end >= 0 {
			s = s[:end]
		}
		return strings.TrimSpace(s)
	}
	// 处理 ``` ... ``` 代码块.
	if idx := strings.Index(s, "```"); idx >= 0 {
		s = s[idx+3:]
		if end := strings.Index(s, "```"); end >= 0 {
			s = s[:end]
		}
		return strings.TrimSpace(s)
	}
	return s
}

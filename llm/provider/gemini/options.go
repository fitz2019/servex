package gemini

import "net/http"

const defaultBaseURL = "https://generativelanguage.googleapis.com"

// Option Gemini 客户端选项.
type Option func(*Client)

// WithBaseURL 设置 API 基础 URL.
func WithBaseURL(url string) Option {
	return func(c *Client) { c.baseURL = url }
}

// WithModel 设置默认模型名称.
func WithModel(model string) Option {
	return func(c *Client) { c.model = model }
}

// WithHTTPClient 设置自定义 HTTP 客户端.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) { c.httpClient = hc }
}

// WithEmbeddingModel 设置默认嵌入模型.
func WithEmbeddingModel(model string) Option {
	return func(c *Client) { c.embeddingModel = model }
}

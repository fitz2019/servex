package github

import "net/http"

type options struct {
	clientID     string
	clientSecret string
	redirectURL  string
	scopes       []string
	httpClient   *http.Client
}

// Option 配置 GitHub OAuth2 Provider 的选项函数.
type Option func(*options)

// WithClientID 设置 GitHub OAuth2 客户端 ID.
func WithClientID(id string) Option { return func(o *options) { o.clientID = id } }

// WithClientSecret 设置 GitHub OAuth2 客户端密钥.
func WithClientSecret(s string) Option { return func(o *options) { o.clientSecret = s } }

// WithRedirectURL 设置 OAuth2 回调地址.
func WithRedirectURL(url string) Option { return func(o *options) { o.redirectURL = url } }

// WithScopes 设置请求的权限范围.
func WithScopes(scopes ...string) Option { return func(o *options) { o.scopes = scopes } }

// WithHTTPClient 设置自定义 HTTP 客户端.
func WithHTTPClient(c *http.Client) Option { return func(o *options) { o.httpClient = c } }

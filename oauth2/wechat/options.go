// oauth2/wechat/options.go
package wechat

import "net/http"

type options struct {
	appID      string
	appSecret  string
	httpClient *http.Client
}

type Option func(*options)

func WithAppID(id string) Option           { return func(o *options) { o.appID = id } }
func WithAppSecret(s string) Option        { return func(o *options) { o.appSecret = s } }
func WithHTTPClient(c *http.Client) Option { return func(o *options) { o.httpClient = c } }

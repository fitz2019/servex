// oauth2/github/options.go
package github

import "net/http"

type options struct {
	clientID     string
	clientSecret string
	redirectURL  string
	scopes       []string
	httpClient   *http.Client
}

type Option func(*options)

func WithClientID(id string) Option       { return func(o *options) { o.clientID = id } }
func WithClientSecret(s string) Option    { return func(o *options) { o.clientSecret = s } }
func WithRedirectURL(url string) Option   { return func(o *options) { o.redirectURL = url } }
func WithScopes(scopes ...string) Option  { return func(o *options) { o.scopes = scopes } }
func WithHTTPClient(c *http.Client) Option { return func(o *options) { o.httpClient = c } }

// Package github 实现 GitHub OAuth2 登录.
package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Tsukikage7/servex/oauth2"
)

const (
	defaultAuthURL     = "https://github.com/login/oauth/authorize"
	defaultTokenURL    = "https://github.com/login/oauth/access_token"
	defaultUserInfoURL = "https://api.github.com/user"
)

// Provider 实现 GitHub OAuth2 登录.
type Provider struct {
	opts        options
	authBaseURL string
	tokenURL    string
	userInfoURL string
}

// NewProvider 创建 GitHub OAuth2 Provider 实例.
func NewProvider(opts ...Option) *Provider {
	o := options{
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
	for _, opt := range opts {
		opt(&o)
	}
	return &Provider{
		opts:        o,
		authBaseURL: defaultAuthURL,
		tokenURL:    defaultTokenURL,
		userInfoURL: defaultUserInfoURL,
	}
}

// AuthURL 生成 GitHub OAuth2 授权跳转链接.
func (p *Provider) AuthURL(state string, opts ...oauth2.AuthURLOption) string {
	extra := oauth2.ApplyAuthURLOptions(opts)

	params := url.Values{
		"client_id":    {p.opts.clientID},
		"redirect_uri": {p.opts.redirectURL},
		"state":        {state},
	}

	scopes := append(p.opts.scopes, extra.Scopes...)
	if len(scopes) > 0 {
		params.Set("scope", strings.Join(scopes, " "))
	}

	return p.authBaseURL + "?" + params.Encode()
}

// Exchange 使用授权码换取访问令牌.
func (p *Provider) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	if code == "" {
		return nil, oauth2.ErrInvalidCode
	}

	data := url.Values{
		"client_id":     {p.opts.clientID},
		"client_secret": {p.opts.clientSecret},
		"code":          {code},
		"redirect_uri":  {p.opts.redirectURL},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := p.opts.httpClient.Do(req)
	if err != nil {
		return nil, errors.Join(oauth2.ErrExchangeFailed, err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, errors.Join(oauth2.ErrExchangeFailed, err)
	}

	if errMsg, ok := result["error"]; ok {
		return nil, fmt.Errorf("%w: %v", oauth2.ErrExchangeFailed, errMsg)
	}

	token := &oauth2.Token{
		AccessToken: getString(result, "access_token"),
		TokenType:   getString(result, "token_type"),
		Raw:         result,
	}
	if scope := getString(result, "scope"); scope != "" {
		token.Scopes = strings.Split(scope, ",")
	}
	return token, nil
}

// Refresh 刷新访问令牌（GitHub 不支持 refresh token）.
func (p *Provider) Refresh(_ context.Context, _ string) (*oauth2.Token, error) {
	// GitHub OAuth2 不支持 refresh token
	return nil, oauth2.ErrRefreshFailed
}

// UserInfo 获取 GitHub 用户信息.
func (p *Provider) UserInfo(ctx context.Context, token *oauth2.Token) (*oauth2.UserInfo, error) {
	if token == nil || token.AccessToken == "" {
		return nil, oauth2.ErrInvalidToken
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.userInfoURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := p.opts.httpClient.Do(req)
	if err != nil {
		return nil, errors.Join(oauth2.ErrUserInfoFailed, err)
	}
	defer resp.Body.Close()

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, errors.Join(oauth2.ErrUserInfoFailed, err)
	}

	return &oauth2.UserInfo{
		ProviderID: fmt.Sprintf("%v", result["id"]),
		Provider:   "github",
		Name:       getString(result, "login"),
		Email:      getString(result, "email"),
		AvatarURL:  getString(result, "avatar_url"),
		Extra:      result,
	}, nil
}

func getString(m map[string]any, key string) string {
	v, _ := m[key].(string)
	return v
}

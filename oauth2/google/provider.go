// oauth2/google/provider.go
package google

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
	defaultAuthURL     = "https://accounts.google.com/o/oauth2/v2/auth"
	defaultTokenURL    = "https://oauth2.googleapis.com/token"
	defaultUserInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo"
)

type Provider struct {
	opts        options
	authBaseURL string
	tokenURL    string
	userInfoURL string
}

func NewProvider(opts ...Option) *Provider {
	o := options{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		scopes:     []string{"openid", "profile", "email"},
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

func (p *Provider) AuthURL(state string, opts ...oauth2.AuthURLOption) string {
	extra := oauth2.ApplyAuthURLOptions(opts)
	params := url.Values{
		"client_id":     {p.opts.clientID},
		"redirect_uri":  {p.opts.redirectURL},
		"response_type": {"code"},
		"state":         {state},
		"access_type":   {"offline"},
	}
	scopes := append(p.opts.scopes, extra.Scopes...)
	params.Set("scope", strings.Join(scopes, " "))
	if extra.Prompt != "" {
		params.Set("prompt", extra.Prompt)
	}
	return p.authBaseURL + "?" + params.Encode()
}

func (p *Provider) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	if code == "" {
		return nil, oauth2.ErrInvalidCode
	}
	data := url.Values{
		"client_id":     {p.opts.clientID},
		"client_secret": {p.opts.clientSecret},
		"code":          {code},
		"redirect_uri":  {p.opts.redirectURL},
		"grant_type":    {"authorization_code"},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

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
	if _, ok := result["error"]; ok {
		return nil, fmt.Errorf("%w: %v", oauth2.ErrExchangeFailed, result["error_description"])
	}

	token := &oauth2.Token{
		AccessToken:  getString(result, "access_token"),
		TokenType:    getString(result, "token_type"),
		RefreshToken: getString(result, "refresh_token"),
		Raw:          result,
	}
	if exp, ok := result["expires_in"].(float64); ok {
		token.ExpiresAt = time.Now().Add(time.Duration(exp) * time.Second)
	}
	if scope := getString(result, "scope"); scope != "" {
		token.Scopes = strings.Split(scope, " ")
	}
	return token, nil
}

func (p *Provider) Refresh(ctx context.Context, refreshToken string) (*oauth2.Token, error) {
	if refreshToken == "" {
		return nil, oauth2.ErrRefreshFailed
	}
	data := url.Values{
		"client_id":     {p.opts.clientID},
		"client_secret": {p.opts.clientSecret},
		"refresh_token": {refreshToken},
		"grant_type":    {"refresh_token"},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.opts.httpClient.Do(req)
	if err != nil {
		return nil, errors.Join(oauth2.ErrRefreshFailed, err)
	}
	defer resp.Body.Close()

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, errors.Join(oauth2.ErrRefreshFailed, err)
	}

	token := &oauth2.Token{
		AccessToken:  getString(result, "access_token"),
		TokenType:    getString(result, "token_type"),
		RefreshToken: refreshToken,
		Raw:          result,
	}
	if exp, ok := result["expires_in"].(float64); ok {
		token.ExpiresAt = time.Now().Add(time.Duration(exp) * time.Second)
	}
	return token, nil
}

func (p *Provider) UserInfo(ctx context.Context, token *oauth2.Token) (*oauth2.UserInfo, error) {
	if token == nil || token.AccessToken == "" {
		return nil, oauth2.ErrInvalidToken
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.userInfoURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

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
		ProviderID: getString(result, "id"),
		Provider:   "google",
		Name:       getString(result, "name"),
		Email:      getString(result, "email"),
		AvatarURL:  getString(result, "picture"),
		Extra:      result,
	}, nil
}

func getString(m map[string]any, key string) string {
	v, _ := m[key].(string)
	return v
}

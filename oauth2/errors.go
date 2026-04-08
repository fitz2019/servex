package oauth2

import "errors"

// ErrInvalidState 表示 state 参数无效.
var ErrInvalidState = errors.New("oauth2: state 无效")

// ErrExchangeFailed 表示 code 换取 token 失败.
var ErrExchangeFailed = errors.New("oauth2: code 换取 token 失败")

// ErrRefreshFailed 表示刷新 token 失败.
var ErrRefreshFailed = errors.New("oauth2: 刷新 token 失败")

// ErrUserInfoFailed 表示获取用户信息失败.
var ErrUserInfoFailed = errors.New("oauth2: 获取用户信息失败")

// ErrInvalidCode 表示 code 为空.
var ErrInvalidCode = errors.New("oauth2: code 为空")

// ErrInvalidToken 表示 token 为空.
var ErrInvalidToken = errors.New("oauth2: token 为空")

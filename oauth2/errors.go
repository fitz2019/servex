// oauth2/errors.go
package oauth2

import "errors"

var (
	ErrInvalidState   = errors.New("oauth2: state 无效")
	ErrExchangeFailed = errors.New("oauth2: code 换取 token 失败")
	ErrRefreshFailed  = errors.New("oauth2: 刷新 token 失败")
	ErrUserInfoFailed = errors.New("oauth2: 获取用户信息失败")
	ErrInvalidCode    = errors.New("oauth2: code 为空")
	ErrInvalidToken   = errors.New("oauth2: token 为空")
)

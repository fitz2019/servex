package tenant

import "errors"

var (
	// ErrMissingToken 缺少租户标识.
	ErrMissingToken = errors.New("tenant: 缺少租户标识")
	// ErrTenantNotFound 租户不存在.
	ErrTenantNotFound = errors.New("tenant: 租户不存在")
	// ErrTenantDisabled 租户已禁用.
	ErrTenantDisabled = errors.New("tenant: 租户已禁用")
)

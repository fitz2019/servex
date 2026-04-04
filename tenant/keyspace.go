package tenant

import (
	"context"
	"strings"
)

// PrefixKey 为 key 添加租户前缀: "{tenantID}/{key}".
// 无租户时原样返回 key.
func PrefixKey(ctx context.Context, key string) string {
	id := ID(ctx)
	if id == "" {
		return key
	}
	return id + "/" + key
}

// StripPrefix 从带前缀的 key 中分离 tenantID 和原始 key.
// 返回 ok=false 表示 key 中不包含前缀.
func StripPrefix(prefixedKey string) (tenantID, key string, ok bool) {
	tenantID, key, ok = strings.Cut(prefixedKey, "/")
	if !ok {
		return "", prefixedKey, false
	}
	return tenantID, key, true
}

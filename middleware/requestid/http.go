package requestid

import (
	"net/http"
)

// HTTPMiddleware 创建 HTTP Request ID 中间件.
// 优先从请求头读取已有 ID，若不存在则生成新 ID，
// 并将 ID 注入 context 和写入响应头.
func HTTPMiddleware(opts ...Option) func(http.Handler) http.Handler {
	o := defaultOptions(opts)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := resolveID(r.Header.Get(o.Header), o.Generator)

			// 注入 context
			ctx := newContextWithID(r.Context(), id)

			// 写入响应头透传
			w.Header().Set(o.Header, id)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

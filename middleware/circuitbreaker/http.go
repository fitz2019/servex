package circuitbreaker

import (
	"net/http"
)

// HTTPMiddleware 创建 HTTP 熔断器中间件.
//
// 熔断器开路时返回 503 Service Unavailable.
func HTTPMiddleware(cb CircuitBreaker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := cb.Execute(r.Context(), func() error {
				next.ServeHTTP(w, r)
				return nil
			})
			if err != nil {
				http.Error(w, "服务暂时不可用，请稍后重试", http.StatusServiceUnavailable)
			}
		})
	}
}

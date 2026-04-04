package cqrs

import "context"

// QueryHandler 查询处理器接口.
type QueryHandler[Q, R any] interface {
	Handle(ctx context.Context, query Q) (R, error)
}

// QueryMiddleware 查询处理器中间件.
type QueryMiddleware[Q, R any] func(QueryHandler[Q, R]) QueryHandler[Q, R]

// queryHandlerFunc 将函数适配为 QueryHandler.
type queryHandlerFunc[Q, R any] struct {
	fn func(ctx context.Context, query Q) (R, error)
}

func (h *queryHandlerFunc[Q, R]) Handle(ctx context.Context, query Q) (R, error) {
	return h.fn(ctx, query)
}

// ChainQuery 将中间件链应用到查询处理器上.
//
// 中间件按参数顺序从外到内执行，即 mws[0] 最先执行.
func ChainQuery[Q, R any](handler QueryHandler[Q, R], mws ...QueryMiddleware[Q, R]) QueryHandler[Q, R] {
	for i := len(mws) - 1; i >= 0; i-- {
		handler = mws[i](handler)
	}
	return handler
}

// ApplyQueryHandler 应用查询处理器.
func ApplyQueryHandler[Q, R any](ctx context.Context, query Q, handler QueryHandler[Q, R]) (R, error) {
	return handler.Handle(ctx, query)
}

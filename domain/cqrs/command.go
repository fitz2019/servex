// Package cqrs 实现CQRS模式的命令和查询处理.
package cqrs

import "context"

// CommandHandler 命令处理器接口.
type CommandHandler[C, R any] interface {
	Handle(ctx context.Context, cmd C) (C, R, error)
}

// CommandMiddleware 命令处理器中间件.
type CommandMiddleware[C, R any] func(CommandHandler[C, R]) CommandHandler[C, R]

// commandHandlerFunc 将函数适配为 CommandHandler.
type commandHandlerFunc[C, R any] struct {
	fn func(ctx context.Context, cmd C) (C, R, error)
}

func (h *commandHandlerFunc[C, R]) Handle(ctx context.Context, cmd C) (C, R, error) {
	return h.fn(ctx, cmd)
}

// ChainCommand 将中间件链应用到命令处理器上.
//
// 中间件按参数顺序从外到内执行，即 mws[0] 最先执行.
func ChainCommand[C, R any](handler CommandHandler[C, R], mws ...CommandMiddleware[C, R]) CommandHandler[C, R] {
	for i := len(mws) - 1; i >= 0; i-- {
		handler = mws[i](handler)
	}
	return handler
}

// ApplyCommand 应用命令处理器.
func ApplyCommand[C, R any](ctx context.Context, cmd C, handler CommandHandler[C, R]) (C, R, error) {
	return handler.Handle(ctx, cmd)
}

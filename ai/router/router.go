// Package router 提供多 Provider 路由器，按模型名称将请求转发到对应的 ChatModel 实现.
package router

import (
	"context"
	"slices"

	"github.com/Tsukikage7/servex/ai"
)

// 编译期接口断言.
var _ ai.ChatModel = (*Router)(nil)

// Route 路由条目：将一组模型名映射到一个 ChatModel.
type Route struct {
	Models []string     // 此路由支持的模型名列表（精确匹配）
	Model  ai.ChatModel // 对应的 Provider 客户端
}

// Router 多 Provider 路由器，实现 ai.ChatModel 接口.
// 根据 WithModel() 选项中的 model 名称，将请求转发到匹配的 Provider.
// 无匹配时使用 fallback.
type Router struct {
	routes   []Route
	fallback ai.ChatModel
}

// New 创建路由器.
// fallback 为必填，当 model 未命中任何路由时使用.
// routes 按顺序匹配，第一个命中的路由生效.
func New(fallback ai.ChatModel, routes ...Route) *Router {
	return &Router{routes: routes, fallback: fallback}
}

// Generate 路由到匹配的 Provider 执行非流式生成.
func (r *Router) Generate(ctx context.Context, messages []ai.Message, opts ...ai.CallOption) (*ai.ChatResponse, error) {
	return r.pick(opts).Generate(ctx, messages, opts...)
}

// Stream 路由到匹配的 Provider 执行流式生成.
func (r *Router) Stream(ctx context.Context, messages []ai.Message, opts ...ai.CallOption) (ai.StreamReader, error) {
	return r.pick(opts).Stream(ctx, messages, opts...)
}

// pick 根据 CallOption 中的 model 名称选取目标 ChatModel.
func (r *Router) pick(opts []ai.CallOption) ai.ChatModel {
	model := ai.ApplyOptions(opts).Model
	if model == "" {
		return r.fallback
	}
	for _, route := range r.routes {
		if slices.Contains(route.Models, model) {
			return route.Model
		}
	}
	return r.fallback
}

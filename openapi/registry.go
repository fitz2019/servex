// openapi/registry.go
package openapi

import "strings"

// Registry 收集 API 端点信息，构建 OpenAPI Spec。
type Registry struct {
	operations []*Operation
	opts       registryOptions
}

func NewRegistry(opts ...RegistryOption) *Registry {
	var o registryOptions
	o.version = "0.0.1"
	for _, opt := range opts {
		opt(&o)
	}
	return &Registry{opts: o}
}

// Add 注册一个 API 端点。
func (r *Registry) Add(ops ...*Operation) {
	r.operations = append(r.operations, ops...)
}

// Build 构建完整的 OpenAPI Spec。
func (r *Registry) Build() *Spec {
	spec := &Spec{
		OpenAPI: "3.0.3",
		Info: Info{
			Title:       r.opts.title,
			Version:     r.opts.version,
			Description: r.opts.description,
		},
		Servers: r.opts.servers,
		Paths:   make(map[string]*PathItem),
	}

	for _, op := range r.operations {
		item, ok := spec.Paths[op.Path]
		if !ok {
			item = &PathItem{}
			spec.Paths[op.Path] = item
		}

		opSpec := r.buildOperationSpec(op)
		switch strings.ToUpper(op.Method) {
		case "GET":
			item.GET = opSpec
		case "POST":
			item.POST = opSpec
		case "PUT":
			item.PUT = opSpec
		case "DELETE":
			item.DELETE = opSpec
		case "PATCH":
			item.PATCH = opSpec
		}
	}

	return spec
}

func (r *Registry) buildOperationSpec(op *Operation) *OperationSpec {
	spec := &OperationSpec{
		Summary:     op.Summary,
		Description: op.Description,
		OperationID: op.OperationID,
		Tags:        op.Tags,
		Deprecated:  op.IsDeprecated,
		Responses:   make(map[string]*Response),
	}

	// Request body
	if op.RequestType != nil {
		schema := SchemaFrom(op.RequestType)
		spec.RequestBody = &RequestBody{
			Required: true,
			Content: map[string]MediaType{
				"application/json": {Schema: schema},
			},
		}
	}

	// Response
	if op.ResponseType != nil {
		schema := SchemaFrom(op.ResponseType)
		spec.Responses["200"] = &Response{
			Description: "成功",
			Content: map[string]MediaType{
				"application/json": {Schema: schema},
			},
		}
	} else {
		spec.Responses["200"] = &Response{Description: "成功"}
	}

	return spec
}

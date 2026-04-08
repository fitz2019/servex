// openapi/operation.go
package openapi

// Operation 描述一个 API 端点（用户注册时传入的数据）。
type Operation struct {
	Method       string
	Path         string
	Summary      string
	Description  string
	Tags         []string
	OperationID  string
	IsDeprecated bool
	RequestType  any
	ResponseType any
	ErrorTypes   []any
}

// Builder 链式构建 Operation。
type Builder struct {
	op Operation
}

func newBuilder(method, path string) *Builder {
	return &Builder{op: Operation{Method: method, Path: path}}
}

func GET(path string) *Builder    { return newBuilder("GET", path) }
func POST(path string) *Builder   { return newBuilder("POST", path) }
func PUT(path string) *Builder    { return newBuilder("PUT", path) }
func DELETE(path string) *Builder { return newBuilder("DELETE", path) }
func PATCH(path string) *Builder  { return newBuilder("PATCH", path) }

func (b *Builder) Summary(s string) *Builder      { b.op.Summary = s; return b }
func (b *Builder) Description(s string) *Builder  { b.op.Description = s; return b }
func (b *Builder) Tags(tags ...string) *Builder   { b.op.Tags = append(b.op.Tags, tags...); return b }
func (b *Builder) OperationID(id string) *Builder { b.op.OperationID = id; return b }
func (b *Builder) Deprecated(d bool) *Builder     { b.op.IsDeprecated = d; return b }
func (b *Builder) Request(v any) *Builder         { b.op.RequestType = v; return b }
func (b *Builder) Response(v any) *Builder        { b.op.ResponseType = v; return b }
func (b *Builder) Errors(types ...any) *Builder   { b.op.ErrorTypes = types; return b }
func (b *Builder) Build() *Operation              { op := b.op; return &op }

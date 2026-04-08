package openapi

type registryOptions struct {
	title       string
	version     string
	description string
	servers     []Server
}

// RegistryOption 定义 Registry 的可选配置函数.
type RegistryOption func(*registryOptions)

// WithInfo 设置 OpenAPI 文档的标题、版本和描述.
func WithInfo(title, version, description string) RegistryOption {
	return func(o *registryOptions) {
		o.title = title
		o.version = version
		o.description = description
	}
}

// WithServer 添加一个 API 服务器地址.
func WithServer(url string, desc ...string) RegistryOption {
	return func(o *registryOptions) {
		s := Server{URL: url}
		if len(desc) > 0 {
			s.Description = desc[0]
		}
		o.servers = append(o.servers, s)
	}
}

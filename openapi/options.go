// openapi/options.go
package openapi

type registryOptions struct {
	title       string
	version     string
	description string
	servers     []Server
}

type RegistryOption func(*registryOptions)

func WithInfo(title, version, description string) RegistryOption {
	return func(o *registryOptions) {
		o.title = title
		o.version = version
		o.description = description
	}
}

func WithServer(url string, desc ...string) RegistryOption {
	return func(o *registryOptions) {
		s := Server{URL: url}
		if len(desc) > 0 {
			s.Description = desc[0]
		}
		o.servers = append(o.servers, s)
	}
}

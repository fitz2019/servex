package openapi

import (
	"encoding/json"
	"net/http"

	"gopkg.in/yaml.v3"
)

// ServeJSON 返回一个 http.Handler，输出 OpenAPI JSON.
func (r *Registry) ServeJSON() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		spec := r.Build()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(spec)
	})
}

// ServeYAML 返回一个 http.Handler，输出 OpenAPI YAML.
func (r *Registry) ServeYAML() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		spec := r.Build()
		w.Header().Set("Content-Type", "application/x-yaml")
		yaml.NewEncoder(w).Encode(spec)
	})
}

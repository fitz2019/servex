package testx

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
)

// HTTPTestServer 封装 httptest.Server，提供便捷的请求方法.
type HTTPTestServer struct {
	*httptest.Server
}

// NewHTTPTestServer 创建 HTTP 测试服务器，支持可选的中间件链.
func NewHTTPTestServer(handler http.Handler, middlewares ...func(http.Handler) http.Handler) *HTTPTestServer {
	// 按逆序包装中间件，使得第一个中间件在最外层执行.
	h := handler
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return &HTTPTestServer{
		Server: httptest.NewServer(h),
	}
}

// Do 执行 HTTP 请求，失败时 panic.
func (s *HTTPTestServer) Do(req *http.Request) *http.Response {
	resp, err := s.Client().Do(req)
	if err != nil {
		panic("testx: HTTP 请求执行失败: " + err.Error())
	}
	return resp
}

// Get 发送 GET 请求到指定路径.
func (s *HTTPTestServer) Get(path string) *http.Response {
	req, err := http.NewRequest(http.MethodGet, s.URL+path, nil)
	if err != nil {
		panic("testx: 构造 GET 请求失败: " + err.Error())
	}
	return s.Do(req)
}

// PostJSON 发送 JSON POST 请求到指定路径.
func (s *HTTPTestServer) PostJSON(path string, body any) *http.Response {
	data, err := json.Marshal(body)
	if err != nil {
		panic("testx: JSON 序列化失败: " + err.Error())
	}
	req, err := http.NewRequest(http.MethodPost, s.URL+path, bytes.NewReader(data))
	if err != nil {
		panic("testx: 构造 POST 请求失败: " + err.Error())
	}
	req.Header.Set("Content-Type", "application/json")
	return s.Do(req)
}

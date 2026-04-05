package bodylimit

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPMiddleware_UnderLimit(t *testing.T) {
	handler := HTTPMiddleware(1024)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		assert.Equal(t, 100, len(body))
		w.WriteHeader(http.StatusOK)
	}))

	body := bytes.Repeat([]byte("a"), 100)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/octet-stream")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHTTPMiddleware_OverLimitContentLength(t *testing.T) {
	handler := HTTPMiddleware(100)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("不应到达 handler")
	}))

	body := bytes.Repeat([]byte("a"), 200)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.ContentLength = 200
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
}

func TestHTTPMiddleware_OverLimitMaxBytesReader(t *testing.T) {
	handler := HTTPMiddleware(100)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := io.ReadAll(r.Body)
		// MaxBytesReader 在超出限制时返回错误
		assert.Error(t, err)
	}))

	// 不设置 ContentLength 以绕过快速检查，测试 MaxBytesReader
	body := bytes.Repeat([]byte("a"), 200)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.ContentLength = -1
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
}

func TestHTTPMiddleware_ExactLimit(t *testing.T) {
	handler := HTTPMiddleware(100)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		assert.Equal(t, 100, len(body))
		w.WriteHeader(http.StatusOK)
	}))

	body := bytes.Repeat([]byte("a"), 100)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestParseLimit(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int64
		wantErr bool
	}{
		{name: "字节", input: "100B", want: 100},
		{name: "KB", input: "1KB", want: 1024},
		{name: "MB", input: "1MB", want: 1024 * 1024},
		{name: "GB", input: "1GB", want: 1024 * 1024 * 1024},
		{name: "TB", input: "1TB", want: 1024 * 1024 * 1024 * 1024},
		{name: "小写", input: "512kb", want: 512 * 1024},
		{name: "带空格", input: " 10 MB ", want: 10 * 1024 * 1024},
		{name: "小数", input: "1.5MB", want: int64(1.5 * 1024 * 1024)},
		{name: "无效格式", input: "abc", wantErr: true},
		{name: "空字符串", input: "", wantErr: true},
		{name: "无单位", input: "100", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseLimit(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

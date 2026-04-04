// openapi/handler_test.go
package openapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestServeJSON(t *testing.T) {
	reg := NewRegistry(WithInfo("Test", "1.0.0", ""))
	reg.Add(GET("/ping").Summary("健康检查").Build())

	handler := reg.ServeJSON()
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/openapi.json", nil)
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("content-type = %s", ct)
	}

	var spec Spec
	if err := json.Unmarshal(w.Body.Bytes(), &spec); err != nil {
		t.Fatal(err)
	}
	if spec.Info.Title != "Test" {
		t.Errorf("title = %s", spec.Info.Title)
	}
}

func TestServeYAML(t *testing.T) {
	reg := NewRegistry(WithInfo("Test", "1.0.0", ""))
	reg.Add(GET("/ping").Summary("健康检查").Build())

	handler := reg.ServeYAML()
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/openapi.yaml", nil)
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "openapi:") {
		t.Error("YAML should contain openapi key")
	}
	if !strings.Contains(body, "健康检查") {
		t.Error("YAML should contain summary")
	}
}

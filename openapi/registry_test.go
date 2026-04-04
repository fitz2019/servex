// openapi/registry_test.go
package openapi

import (
	"encoding/json"
	"testing"
)

type orderReq struct {
	UserID string  `json:"user_id" validate:"required" description:"用户ID"`
	Amount float64 `json:"amount" validate:"required,min=0.01" description:"金额"`
}

type orderResp struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

func TestRegistry_Build(t *testing.T) {
	reg := NewRegistry(
		WithInfo("Test API", "1.0.0", "测试"),
		WithServer("https://api.example.com"),
	)

	reg.Add(POST("/orders").
		Summary("创建订单").
		Tags("orders").
		Request(orderReq{}).
		Response(orderResp{}).
		Build(),
	)

	reg.Add(GET("/orders/{id}").
		Summary("查询订单").
		Tags("orders").
		Response(orderResp{}).
		Build(),
	)

	spec := reg.Build()

	if spec.OpenAPI != "3.0.3" {
		t.Errorf("openapi = %s", spec.OpenAPI)
	}
	if spec.Info.Title != "Test API" {
		t.Errorf("title = %s", spec.Info.Title)
	}
	if len(spec.Servers) != 1 {
		t.Fatalf("servers = %d", len(spec.Servers))
	}
	if len(spec.Paths) != 2 {
		t.Fatalf("paths = %d", len(spec.Paths))
	}

	orderPath := spec.Paths["/orders"]
	if orderPath == nil || orderPath.POST == nil {
		t.Fatal("/orders POST should exist")
	}
	if orderPath.POST.Summary != "创建订单" {
		t.Errorf("summary = %s", orderPath.POST.Summary)
	}
	if orderPath.POST.RequestBody == nil {
		t.Fatal("request body should exist")
	}

	getPath := spec.Paths["/orders/{id}"]
	if getPath == nil || getPath.GET == nil {
		t.Fatal("/orders/{id} GET should exist")
	}

	// 验证可以序列化为 JSON
	data, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Error("JSON output should not be empty")
	}
}

func TestRegistry_Empty(t *testing.T) {
	reg := NewRegistry(WithInfo("Empty", "0.0.1", ""))
	spec := reg.Build()
	if len(spec.Paths) != 0 {
		t.Error("empty registry should have no paths")
	}
}

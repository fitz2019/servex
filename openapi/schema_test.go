// openapi/schema_test.go
package openapi

import (
	"testing"
	"time"
)

type testRequest struct {
	UserID string  `json:"user_id" validate:"required" description:"用户ID"`
	Amount float64 `json:"amount" validate:"required,min=0.01" description:"金额"`
	Remark string  `json:"remark,omitempty" description:"备注"`
	Count  int     `json:"count" validate:"min=1,max=100"`
}

func TestSchemaFrom_Struct(t *testing.T) {
	s := SchemaFrom(testRequest{})

	if s.Type != "object" {
		t.Errorf("type = %s, want object", s.Type)
	}

	// 检查 required 字段
	requiredSet := make(map[string]bool)
	for _, r := range s.Required {
		requiredSet[r] = true
	}
	if !requiredSet["user_id"] {
		t.Error("user_id should be required")
	}
	if !requiredSet["amount"] {
		t.Error("amount should be required")
	}
	if requiredSet["remark"] {
		t.Error("remark should not be required")
	}

	// 检查 properties
	if s.Properties["user_id"].Type != "string" {
		t.Errorf("user_id type = %s", s.Properties["user_id"].Type)
	}
	if s.Properties["user_id"].Description != "用户ID" {
		t.Errorf("user_id description = %s", s.Properties["user_id"].Description)
	}
	if s.Properties["amount"].Type != "number" {
		t.Errorf("amount type = %s", s.Properties["amount"].Type)
	}
	if s.Properties["amount"].Minimum == nil || *s.Properties["amount"].Minimum != 0.01 {
		t.Error("amount should have minimum 0.01")
	}
	if s.Properties["count"].Minimum == nil || *s.Properties["count"].Minimum != 1 {
		t.Error("count should have minimum 1")
	}
	if s.Properties["count"].Maximum == nil || *s.Properties["count"].Maximum != 100 {
		t.Error("count should have maximum 100")
	}
}

func TestSchemaFrom_Primitives(t *testing.T) {
	tests := []struct {
		name string
		val  any
		want string
	}{
		{"string", "", "string"},
		{"int", 0, "integer"},
		{"float64", 0.0, "number"},
		{"bool", false, "boolean"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := SchemaFrom(tt.val)
			if s.Type != tt.want {
				t.Errorf("type = %s, want %s", s.Type, tt.want)
			}
		})
	}
}

func TestSchemaFrom_Slice(t *testing.T) {
	s := SchemaFrom([]string{})
	if s.Type != "array" {
		t.Errorf("type = %s, want array", s.Type)
	}
	if s.Items == nil || s.Items.Type != "string" {
		t.Error("items should be string type")
	}
}

type nestedStruct struct {
	Name    string `json:"name"`
	Address struct {
		City string `json:"city"`
	} `json:"address"`
}

func TestSchemaFrom_Nested(t *testing.T) {
	s := SchemaFrom(nestedStruct{})
	if s.Properties["address"] == nil {
		t.Fatal("address property should exist")
	}
	if s.Properties["address"].Properties["city"] == nil {
		t.Fatal("address.city property should exist")
	}
}

func TestSchemaFrom_Time(t *testing.T) {
	s := SchemaFrom(time.Time{})
	if s.Type != "string" || s.Format != "date-time" {
		t.Errorf("type=%s format=%s, want string date-time", s.Type, s.Format)
	}
}

func TestSchemaFrom_Pointer(t *testing.T) {
	type withPtr struct {
		Name *string `json:"name"`
	}
	s := SchemaFrom(withPtr{})
	if s.Properties["name"].Type != "string" {
		t.Errorf("pointer string type = %s", s.Properties["name"].Type)
	}
}

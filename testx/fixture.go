package testx

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

// update 控制是否更新 golden 文件的标志.
var update = flag.Bool("update", false, "update golden files")

// LoadJSON 从文件加载 JSON 并反序列化为指定类型.
func LoadJSON[T any](t *testing.T, path string) T {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("testx: 读取 JSON 文件 %s 失败: %v", path, err)
	}
	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		t.Fatalf("testx: 解析 JSON 文件 %s 失败: %v", path, err)
	}
	return v
}

// LoadYAML 从文件加载 YAML 并反序列化为指定类型.
func LoadYAML[T any](t *testing.T, path string) T {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("testx: 读取 YAML 文件 %s 失败: %v", path, err)
	}
	var v T
	if err := yaml.Unmarshal(data, &v); err != nil {
		t.Fatalf("testx: 解析 YAML 文件 %s 失败: %v", path, err)
	}
	return v
}

// Golden 对比 actual 与 golden 文件内容。当 -update 标志启用时，写入新的 golden 文件.
// golden 文件路径格式: testdata/<name>.golden
func Golden(t *testing.T, name string, actual []byte) {
	t.Helper()

	goldenPath := filepath.Join("testdata", name+".golden")

	if *update {
		dir := filepath.Dir(goldenPath)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("testx: 创建 golden 文件目录失败: %v", err)
		}
		if err := os.WriteFile(goldenPath, actual, 0o644); err != nil {
			t.Fatalf("testx: 写入 golden 文件 %s 失败: %v", goldenPath, err)
		}
		return
	}

	expected, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("testx: 读取 golden 文件 %s 失败（使用 -update 生成）: %v", goldenPath, err)
	}

	if string(expected) != string(actual) {
		t.Errorf("testx: golden 文件内容不匹配\n期望:\n%s\n实际:\n%s", string(expected), string(actual))
	}
}

// GoldenJSON 将 actual 序列化为格式化 JSON 后与 golden 文件对比.
func GoldenJSON(t *testing.T, name string, actual any) {
	t.Helper()
	data, err := json.MarshalIndent(actual, "", "  ")
	if err != nil {
		t.Fatalf("testx: JSON 序列化失败: %v", err)
	}
	data = append(data, '\n')
	Golden(t, name, data)
}

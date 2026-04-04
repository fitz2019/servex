package testx

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/Tsukikage7/servex/observability/logger"
)

func TestNopLogger(t *testing.T) {
	l := NopLogger()
	// 确保所有方法都能正常调用不会 panic.
	l.Debug("debug")
	l.Debugf("debug %s", "msg")
	l.Info("info")
	l.Infof("info %s", "msg")
	l.Warn("warn")
	l.Warnf("warn %s", "msg")
	l.Error("error")
	l.Errorf("error %s", "msg")

	// With 应返回同类型 logger.
	l2 := l.With(logger.Field{Key: "k", Value: "v"})
	if l2 == nil {
		t.Fatal("With 返回了 nil")
	}
	l2.Info("with field")

	// Sync 和 Close 应返回 nil.
	if err := l.Sync(); err != nil {
		t.Fatalf("Sync 返回了非 nil 错误: %v", err)
	}
	if err := l.Close(); err != nil {
		t.Fatalf("Close 返回了非 nil 错误: %v", err)
	}
}

func TestTestLogger(t *testing.T) {
	l := TestLogger(t)
	// 验证所有方法可以正常执行，输出会通过 t.Log 记录.
	l.Debug("debug message")
	l.Debugf("debug %s", "formatted")
	l.Info("info message")
	l.Infof("info %d", 42)
	l.Warn("warn message")
	l.Error("error message")

	l2 := l.With(logger.Field{Key: "request_id", Value: "abc123"})
	l2.Info("with field message")
}

func TestLoadJSON(t *testing.T) {
	// 创建临时 JSON 文件.
	type sample struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")

	data, _ := json.Marshal(sample{Name: "hello", Value: 42})
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("写入临时文件失败: %v", err)
	}

	result := LoadJSON[sample](t, path)
	if result.Name != "hello" {
		t.Errorf("Name 不匹配: 期望 %q, 实际 %q", "hello", result.Name)
	}
	if result.Value != 42 {
		t.Errorf("Value 不匹配: 期望 %d, 实际 %d", 42, result.Value)
	}
}

func TestGolden(t *testing.T) {
	// 使用临时目录替代 testdata 目录进行测试.
	dir := t.TempDir()
	goldenDir := filepath.Join(dir, "testdata")
	if err := os.MkdirAll(goldenDir, 0o755); err != nil {
		t.Fatal(err)
	}

	goldenFile := filepath.Join(goldenDir, "sample.golden")
	expected := []byte("hello golden\n")
	if err := os.WriteFile(goldenFile, expected, 0o644); err != nil {
		t.Fatal(err)
	}

	// 保存并切换工作目录到临时目录.
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(orig) })

	// 匹配场景.
	Golden(t, "sample", expected)
}

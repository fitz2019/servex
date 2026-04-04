package notify

import (
	"embed"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

//go:embed testdata/templates
var testFS embed.FS

func TestTemplateEngine_RenderNotFound(t *testing.T) {
	eng := NewTemplateEngine()
	_, err := eng.Render("nonexistent", nil)
	if !errors.Is(err, ErrTemplateNotFound) {
		t.Errorf("got %v, want ErrTemplateNotFound", err)
	}
}

func TestTemplateEngine_WithTemplateDir(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "welcome.html"), []byte("<h1>Hello, {{.Name}}!</h1>"), 0o644); err != nil {
		t.Fatal(err)
	}

	eng := NewTemplateEngine(WithTemplateDir(dir))
	got, err := eng.Render("welcome.html", map[string]any{"Name": "Alice"})
	if err != nil {
		t.Fatal(err)
	}
	if want := "<h1>Hello, Alice!</h1>"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestTemplateEngine_WithTemplateFS(t *testing.T) {
	eng := NewTemplateEngine(WithTemplateFS(testFS, "testdata/templates"))
	got, err := eng.Render("greeting.html", map[string]any{"User": "Bob"})
	if err != nil {
		t.Fatal(err)
	}
	if got == "" {
		t.Error("expected non-empty rendered output")
	}
}

func TestTemplateEngine_RenderError(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "bad.html"), []byte("{{.Name"), 0o644); err != nil {
		t.Fatal(err)
	}
	eng := NewTemplateEngine(WithTemplateDir(dir))
	_, err := eng.Render("bad.html", nil)
	if err == nil {
		t.Error("expected error for bad template")
	}
}

func TestTemplateEngine_NilData(t *testing.T) {
	dir := t.TempDir()
	content := "<p>Static</p>"
	if err := os.WriteFile(filepath.Join(dir, "static.html"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	eng := NewTemplateEngine(WithTemplateDir(dir))
	got, err := eng.Render("static.html", nil)
	if err != nil {
		t.Fatal(err)
	}
	if got != content {
		t.Errorf("got %q, want %q", got, content)
	}
}

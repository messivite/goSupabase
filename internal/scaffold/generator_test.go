package scaffold

import (
	"os"
	"path/filepath"
	"testing"
)

func TestToSnakeCase(t *testing.T) {
	if got := toSnakeCase("CreateTrack"); got != "create_track" {
		t.Fatalf("toSnakeCase(CreateTrack) = %q", got)
	}
}

func TestRenderTemplate(t *testing.T) {
	out, err := renderTemplate("Hello {{.Name}}", map[string]string{"Name": "goSupaBase"})
	if err != nil {
		t.Fatalf("renderTemplate error: %v", err)
	}
	if out != "Hello goSupaBase" {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestScaffoldNewCreatesCoreFiles(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "my-api")
	if err := ScaffoldNew(target, "github.com/example/my-api"); err != nil {
		t.Fatalf("ScaffoldNew error: %v", err)
	}

	paths := []string{
		filepath.Join(target, "go.mod"),
		filepath.Join(target, "api.yaml"),
		filepath.Join(target, ".env.example"),
		filepath.Join(target, "cmd/server/main.go"),
	}
	for _, p := range paths {
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("expected scaffolded file %s: %v", p, err)
		}
	}
}

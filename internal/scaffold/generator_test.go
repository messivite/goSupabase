package scaffold

import (
	"os"
	"path/filepath"
	"strings"
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

func TestDetectModuleFallbackWhenNoGoMod(t *testing.T) {
	orig, _ := os.Getwd()
	dir := t.TempDir()
	_ = os.Chdir(dir)
	defer os.Chdir(orig)

	got := detectModule()
	if got != "github.com/example/myapp" {
		t.Fatalf("detectModule fallback = %q", got)
	}
}

func TestGenerateHandlersOnly(t *testing.T) {
	orig, _ := os.Getwd()
	dir := t.TempDir()
	_ = os.Chdir(dir)
	defer os.Chdir(orig)

	api := `version: "1"
basePath: /api
endpoints:
  - method: GET
    path: /health
    handler: Health
    auth: false
`
	if err := os.WriteFile("api.yaml", []byte(api), 0644); err != nil {
		t.Fatal(err)
	}
	err := Generate("api.yaml", GenerateOptions{
		HandlersDir:  "handlers",
		ServerDir:    "server",
		HandlersOnly: true,
		Module:       "github.com/example/test",
	})
	if err != nil {
		t.Fatalf("Generate handlers-only: %v", err)
	}
	if _, err := os.Stat(filepath.Join("handlers", "health.go")); err != nil {
		t.Fatalf("expected generated handler file: %v", err)
	}
	if _, err := os.Stat(filepath.Join("server", "server.go")); err == nil {
		t.Fatal("server.go should not be generated in handlers-only mode")
	}
}

func TestGenerateSkipsExistingHandler(t *testing.T) {
	orig, _ := os.Getwd()
	dir := t.TempDir()
	_ = os.Chdir(dir)
	defer os.Chdir(orig)

	api := `version: "1"
basePath: /api
endpoints:
  - method: GET
    path: /health
    handler: Health
    auth: false
`
	if err := os.WriteFile("api.yaml", []byte(api), 0644); err != nil {
		t.Fatal(err)
	}
	goMod := "module github.com/example/test\n\ngo 1.22\n"
	if err := os.WriteFile("go.mod", []byte(goMod), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll("middleware", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join("middleware", "stub.go"), []byte("package middleware\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join("internal", "yaml"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join("internal", "yaml", "stub.go"), []byte("package yaml\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll("handlers", 0755); err != nil {
		t.Fatal(err)
	}
	initial := "package handlers\n\nfunc Health() {}\n"
	if err := os.WriteFile(filepath.Join("handlers", "health.go"), []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}
	err := Generate("api.yaml", GenerateOptions{
		HandlersDir: "handlers",
		ServerDir:   "server",
		Module:      "github.com/example/test",
	})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	data, err := os.ReadFile(filepath.Join("handlers", "health.go"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(data)) != strings.TrimSpace(initial) {
		t.Fatal("existing handler file should be kept as-is")
	}
}

func TestGenerateRefusesFullServerWithoutLocalDeps(t *testing.T) {
	orig, _ := os.Getwd()
	dir := t.TempDir()
	_ = os.Chdir(dir)
	defer os.Chdir(orig)

	goMod := "module github.com/example/consumer\n\ngo 1.22\n"
	if err := os.WriteFile("go.mod", []byte(goMod), 0644); err != nil {
		t.Fatal(err)
	}
	api := `version: "1"
basePath: /api
endpoints:
  - method: GET
    path: /health
    handler: Health
    auth: false
`
	if err := os.WriteFile("api.yaml", []byte(api), 0644); err != nil {
		t.Fatal(err)
	}
	err := Generate("api.yaml", GenerateOptions{
		HandlersDir: "handlers",
		ServerDir:   "server",
		Module:      "github.com/example/consumer",
	})
	if err == nil {
		t.Fatal("expected error when middleware/internal/yaml are missing")
	}
	if !strings.Contains(err.Error(), "--handlers-only") {
		t.Fatalf("error should mention --handlers-only: %v", err)
	}
	if _, statErr := os.Stat(filepath.Join("server", "server.go")); statErr == nil {
		t.Fatal("server.go should not be written when deps are missing")
	}
}

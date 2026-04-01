package main

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	yamlcfg "github.com/mustafaaksoy/gosupabase/internal/yaml"
)

func TestSameSnapshot(t *testing.T) {
	now := time.Now()
	a := map[string]time.Time{"a.go": now}
	b := map[string]time.Time{"a.go": now}
	if !sameSnapshot(a, b) {
		t.Fatal("expected snapshots to be equal")
	}
	b["a.go"] = now.Add(time.Second)
	if sameSnapshot(a, b) {
		t.Fatal("expected snapshots to differ")
	}
}

func TestSnapshotWatchedFiles(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.go"), []byte("package main"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "api.yaml"), []byte("version: \"1\""), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("ignore me"), 0644); err != nil {
		t.Fatal(err)
	}

	snap, err := snapshotWatchedFiles(dir)
	if err != nil {
		t.Fatalf("snapshotWatchedFiles error: %v", err)
	}
	if len(snap) < 2 {
		t.Fatalf("expected at least 2 watched files, got %d", len(snap))
	}
}

func TestPromptValidationModeDefaultOnInvalid(t *testing.T) {
	orig := stdinReader
	stdinReader = bufio.NewReader(strings.NewReader("invalid\n"))
	defer func() { stdinReader = orig }()

	got := promptValidationMode("auto")
	if got != "auto" {
		t.Fatalf("promptValidationMode invalid input = %q, want auto", got)
	}
}

func TestPromptConflictChoices(t *testing.T) {
	orig := stdinReader
	defer func() { stdinReader = orig }()

	stdinReader = bufio.NewReader(strings.NewReader("o\n"))
	if got := promptConflict(".env"); got != policyOverwrite {
		t.Fatalf("expected overwrite policy, got %v", got)
	}
	stdinReader = bufio.NewReader(strings.NewReader("m\n"))
	if got := promptConflict(".env"); got != policyMerge {
		t.Fatalf("expected merge policy, got %v", got)
	}
	stdinReader = bufio.NewReader(strings.NewReader("s\n"))
	if got := promptConflict(".env"); got != policySkip {
		t.Fatalf("expected skip policy, got %v", got)
	}
}

func TestWriteAndMergeEnvFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	keys := []string{"PORT", "SUPABASE_URL"}
	initial := map[string]string{"PORT": "8080", "SUPABASE_URL": "https://a.supabase.co"}
	if err := writeEnvFile(path, initial, keys); err != nil {
		t.Fatalf("writeEnvFile: %v", err)
	}

	update := map[string]string{"PORT": "9090", "SUPABASE_URL": "https://b.supabase.co"}
	if err := mergeEnvFile(path, update, keys); err != nil {
		t.Fatalf("mergeEnvFile: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	// Merge keeps existing values.
	if !strings.Contains(s, "PORT=8080") {
		t.Fatalf("expected existing PORT to stay, got: %s", s)
	}
}

func TestSetupFromFileCreatesFiles(t *testing.T) {
	origWD, _ := os.Getwd()
	dir := t.TempDir()
	_ = os.Chdir(dir)
	defer os.Chdir(origWD)

	source := "source.env"
	content := "PORT=9090\nSUPABASE_URL=https://x.supabase.co\nSUPABASE_ANON_KEY=anon\nSUPABASE_JWT_SECRET=secret\nSERVER_DIR=pkg/server\nHANDLERS_DIR=pkg/handlers\n"
	if err := os.WriteFile(source, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	setupFromFile(source)

	if _, err := os.Stat(".env"); err != nil {
		t.Fatalf("expected .env: %v", err)
	}
	if _, err := os.Stat(".gosupabase.yaml"); err != nil {
		t.Fatalf("expected .gosupabase.yaml: %v", err)
	}
}

func TestSetupInteractiveCreatesFiles(t *testing.T) {
	origWD, _ := os.Getwd()
	dir := t.TempDir()
	_ = os.Chdir(dir)
	defer os.Chdir(origWD)

	origReader := stdinReader
	// port, url, anon, secret, validation mode, include service key=no, server dir, handlers dir
	input := "8080\nhttps://x.supabase.co\nanon\nsecret\nauto\nn\nserver\nhandlers\n"
	stdinReader = bufio.NewReader(strings.NewReader(input))
	defer func() { stdinReader = origReader }()

	setupInteractive()

	if _, err := os.Stat(".env"); err != nil {
		t.Fatalf("expected .env: %v", err)
	}
	if _, err := os.Stat(".gosupabase.yaml"); err != nil {
		t.Fatalf("expected .gosupabase.yaml: %v", err)
	}
}

func TestCmdInitCreatesDefaults(t *testing.T) {
	origWD, _ := os.Getwd()
	dir := t.TempDir()
	_ = os.Chdir(dir)
	defer os.Chdir(origWD)

	cmdInit()

	if _, err := os.Stat("api.yaml"); err != nil {
		t.Fatalf("expected api.yaml: %v", err)
	}
	if _, err := os.Stat(".env.example"); err != nil {
		t.Fatalf("expected .env.example: %v", err)
	}
}

func TestCmdAddAndList(t *testing.T) {
	origWD, _ := os.Getwd()
	dir := t.TempDir()
	_ = os.Chdir(dir)
	defer os.Chdir(origWD)

	cfg := &yamlcfg.APIConfig{
		Version:  "1",
		BasePath: "/api",
		Output:   yamlcfg.OutputConfig{ServerDir: "server", HandlersDir: "handlers"},
	}
	if err := yamlcfg.Save("api.yaml", cfg); err != nil {
		t.Fatal(err)
	}

	cmdAdd([]string{"endpoint", "GET /tracks"})

	loaded, err := yamlcfg.Load("api.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Endpoints) != 1 {
		t.Fatalf("expected 1 endpoint, got %d", len(loaded.Endpoints))
	}
	if loaded.Endpoints[0].Handler == "" {
		t.Fatal("expected derived handler name")
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	cmdList()
	_ = w.Close()
	os.Stdout = oldStdout
	var out bytes.Buffer
	_, _ = out.ReadFrom(r)
	_ = r.Close()
	if !strings.Contains(out.String(), "GET") || !strings.Contains(out.String(), "/tracks") {
		t.Fatalf("unexpected list output: %q", out.String())
	}
}

func TestCmdGenHandlersOnly(t *testing.T) {
	origWD, _ := os.Getwd()
	dir := t.TempDir()
	_ = os.Chdir(dir)
	defer os.Chdir(origWD)

	cfg := &yamlcfg.APIConfig{
		Version:  "1",
		BasePath: "/api",
		Output:   yamlcfg.OutputConfig{ServerDir: "server", HandlersDir: "handlers"},
		Endpoints: []yamlcfg.Endpoint{
			{Method: "GET", Path: "/health", Handler: "Health", Auth: false},
		},
	}
	if err := yamlcfg.Save("api.yaml", cfg); err != nil {
		t.Fatal(err)
	}

	cmdGen([]string{"--handlers-only"})

	if _, err := os.Stat(filepath.Join("handlers", "health.go")); err != nil {
		t.Fatalf("expected generated handler file: %v", err)
	}
	if _, err := os.Stat(filepath.Join("server", "server.go")); !os.IsNotExist(err) {
		t.Fatalf("expected no generated server.go in handlers-only mode, got err=%v", err)
	}
}

func TestApplyFileWithPolicySkipExistingEnv(t *testing.T) {
	origWD, _ := os.Getwd()
	dir := t.TempDir()
	_ = os.Chdir(dir)
	defer os.Chdir(origWD)

	if err := os.WriteFile(".env", []byte("PORT=1111\n"), 0644); err != nil {
		t.Fatal(err)
	}
	envMap := map[string]string{"PORT": "8080"}
	keys := []string{"PORT"}

	origReader := stdinReader
	stdinReader = bufio.NewReader(strings.NewReader("s\n"))
	defer func() { stdinReader = origReader }()

	applyFileWithPolicy(".env", envMap, keys, "server", "handlers", true)

	data, err := os.ReadFile(".env")
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(data)) != "PORT=1111" {
		t.Fatalf("expected existing file to stay, got: %q", string(data))
	}
}

func TestApplyFileWithPolicyMergeEnv(t *testing.T) {
	origWD, _ := os.Getwd()
	dir := t.TempDir()
	_ = os.Chdir(dir)
	defer os.Chdir(origWD)

	if err := os.WriteFile(".env", []byte("PORT=1111\n"), 0644); err != nil {
		t.Fatal(err)
	}
	envMap := map[string]string{
		"PORT":         "8080",
		"SUPABASE_URL": "https://x.supabase.co",
	}
	keys := []string{"PORT", "SUPABASE_URL"}

	origReader := stdinReader
	stdinReader = bufio.NewReader(strings.NewReader("m\n"))
	defer func() { stdinReader = origReader }()

	applyFileWithPolicy(".env", envMap, keys, "server", "handlers", true)

	data, err := os.ReadFile(".env")
	if err != nil {
		t.Fatal(err)
	}
	out := string(data)
	if !strings.Contains(out, "PORT=1111") {
		t.Fatalf("expected existing key to remain, got: %q", out)
	}
	if !strings.Contains(out, "SUPABASE_URL=https://x.supabase.co") {
		t.Fatalf("expected missing key to be added, got: %q", out)
	}
}

func TestApplyFileWithPolicyOverwriteYaml(t *testing.T) {
	origWD, _ := os.Getwd()
	dir := t.TempDir()
	_ = os.Chdir(dir)
	defer os.Chdir(origWD)

	if err := os.WriteFile(".gosupabase.yaml", []byte("output:\n  serverDir: old\n  handlersDir: old\n"), 0644); err != nil {
		t.Fatal(err)
	}
	origReader := stdinReader
	stdinReader = bufio.NewReader(strings.NewReader("o\n"))
	defer func() { stdinReader = origReader }()

	applyFileWithPolicy(".gosupabase.yaml", map[string]string{}, nil, "pkg/server", "pkg/handlers", false)

	data, err := os.ReadFile(".gosupabase.yaml")
	if err != nil {
		t.Fatal(err)
	}
	out := string(data)
	if !strings.Contains(out, "serverDir: pkg/server") || !strings.Contains(out, "handlersDir: pkg/handlers") {
		t.Fatalf("unexpected yaml output: %q", out)
	}
}

func TestSnapshotWatchedFilesSkipsIgnoredDirs(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, ".cursor"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "node_modules"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "ok"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".git", "x.go"), []byte("package x"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".cursor", "y.go"), []byte("package y"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "node_modules", "z.go"), []byte("package z"), 0644); err != nil {
		t.Fatal(err)
	}
	okPath := filepath.Join(dir, "ok", "main.go")
	if err := os.WriteFile(okPath, []byte("package main"), 0644); err != nil {
		t.Fatal(err)
	}

	snap, err := snapshotWatchedFiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	for p := range snap {
		if strings.Contains(p, ".git") || strings.Contains(p, ".cursor") || strings.Contains(p, "node_modules") {
			t.Fatalf("ignored directory file should not be watched: %s", p)
		}
	}
	if _, ok := snap[okPath]; !ok {
		t.Fatalf("expected regular go file to be watched: %s", okPath)
	}
}

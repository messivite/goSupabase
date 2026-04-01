package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadEnvDefaults(t *testing.T) {
	os.Unsetenv("PORT")
	cfg := LoadEnv()
	if cfg.Port != "8080" {
		t.Errorf("default port = %q, want %q", cfg.Port, "8080")
	}
}

func TestLoadEnvFromEnvVar(t *testing.T) {
	os.Setenv("PORT", "3000")
	defer os.Unsetenv("PORT")

	cfg := LoadEnv()
	if cfg.Port != "3000" {
		t.Errorf("port = %q, want %q", cfg.Port, "3000")
	}
}

func TestLoadEnvDefaultValidationMode(t *testing.T) {
	os.Unsetenv("SUPABASE_JWT_VALIDATION_MODE")
	cfg := LoadEnv()
	if cfg.SupabaseJWTValidationMode != "auto" {
		t.Errorf("validation mode = %q, want %q", cfg.SupabaseJWTValidationMode, "auto")
	}
}

func TestLoadEnvValidationModeFromEnv(t *testing.T) {
	os.Setenv("SUPABASE_JWT_VALIDATION_MODE", "jwks")
	defer os.Unsetenv("SUPABASE_JWT_VALIDATION_MODE")
	cfg := LoadEnv()
	if cfg.SupabaseJWTValidationMode != "jwks" {
		t.Errorf("validation mode = %q, want %q", cfg.SupabaseJWTValidationMode, "jwks")
	}
}

func TestResolveOutputPathsDefaults(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(orig)

	paths := ResolveOutputPaths("", "")
	if paths.ServerDir != "server" {
		t.Errorf("serverDir = %q, want %q", paths.ServerDir, "server")
	}
	if paths.HandlersDir != "handlers" {
		t.Errorf("handlersDir = %q, want %q", paths.HandlersDir, "handlers")
	}
}

func TestResolveOutputPathsFromAPIYaml(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(orig)

	content := "version: \"1\"\noutput:\n  serverDir: pkg/server\n  handlersDir: pkg/handlers\n"
	os.WriteFile(filepath.Join(dir, "api.yaml"), []byte(content), 0644)

	paths := ResolveOutputPaths("", "")
	if paths.ServerDir != "pkg/server" {
		t.Errorf("serverDir = %q, want %q", paths.ServerDir, "pkg/server")
	}
	if paths.HandlersDir != "pkg/handlers" {
		t.Errorf("handlersDir = %q, want %q", paths.HandlersDir, "pkg/handlers")
	}
}

func TestResolveOutputPathsGosupabaseYAMLOverrides(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(orig)

	os.WriteFile(filepath.Join(dir, "api.yaml"), []byte("output:\n  serverDir: from-api\n  handlersDir: from-api\n"), 0644)
	os.WriteFile(filepath.Join(dir, ".gosupabase.yaml"), []byte("output:\n  serverDir: from-gosupabase\n  handlersDir: from-gosupabase\n"), 0644)

	paths := ResolveOutputPaths("", "")
	if paths.ServerDir != "from-gosupabase" {
		t.Errorf("serverDir = %q, want %q", paths.ServerDir, "from-gosupabase")
	}
}

func TestResolveOutputPathsFlagsOverrideAll(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(orig)

	os.WriteFile(filepath.Join(dir, ".gosupabase.yaml"), []byte("output:\n  serverDir: from-gosupabase\n"), 0644)

	paths := ResolveOutputPaths("flag-server", "flag-handlers")
	if paths.ServerDir != "flag-server" {
		t.Errorf("serverDir = %q, want %q", paths.ServerDir, "flag-server")
	}
	if paths.HandlersDir != "flag-handlers" {
		t.Errorf("handlersDir = %q, want %q", paths.HandlersDir, "flag-handlers")
	}
}

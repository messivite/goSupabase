package deploy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNormalizeProvider(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"vercel", ProviderVercel},
		{"Vercel", ProviderVercel},
		{"v", ProviderVercel},
		{"1", ProviderVercel},
		{"fly", ProviderFly},
		{"fly.io", ProviderFly},
		{"2", ProviderFly},
		{"railway", ProviderRailway},
		{"rail", ProviderRailway},
		{"3", ProviderRailway},
		{"render", ProviderRender},
		{"4", ProviderRender},
		{"none", ProviderNone},
		{"skip", ProviderNone},
		{"local", ProviderNone},
		{"5", ProviderNone},
		{"0", ProviderNone},
		{"", ProviderNone},
		{"nope", ProviderNone},
	}
	for _, tt := range tests {
		if got := NormalizeProvider(tt.in); got != tt.want {
			t.Fatalf("NormalizeProvider(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func testWriteFilesProvider(t *testing.T, provider, wantFile string) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	written, skipped, err := WriteFiles(provider, "test-fly-app", func(string) bool { return true })
	if err != nil {
		t.Fatal(err)
	}
	if len(skipped) != 0 {
		t.Fatalf("skipped = %v", skipped)
	}
	if len(written) < 2 {
		t.Fatalf("expected %s and DEPLOY.md, got %v", wantFile, written)
	}
	for _, name := range []string{wantFile, "DEPLOY.md"} {
		p := filepath.Join(dir, name)
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("missing %s: %v", name, err)
		}
	}
}

func TestWriteFilesVercel(t *testing.T) {
	testWriteFilesProvider(t, ProviderVercel, "vercel.json")
}

func TestWriteFilesFly(t *testing.T) {
	testWriteFilesProvider(t, ProviderFly, "fly.toml")
}

func TestFlyTomlUsesAppName(t *testing.T) {
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	if _, _, err := WriteFiles(ProviderFly, "my-custom-app", func(string) bool { return true }); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(dir, "fly.toml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "my-custom-app") {
		t.Fatalf("fly.toml should contain app name, got: %s", string(b))
	}
}

func TestWriteFilesRailway(t *testing.T) {
	testWriteFilesProvider(t, ProviderRailway, "railway.toml")
}

func TestWriteFilesRender(t *testing.T) {
	testWriteFilesProvider(t, ProviderRender, "render.yaml")
}

func TestWriteFilesNoneWritesNothing(t *testing.T) {
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	written, skipped, err := WriteFiles(ProviderNone, "", func(string) bool { return true })
	if err != nil {
		t.Fatal(err)
	}
	if len(written) != 0 || len(skipped) != 0 {
		t.Fatalf("expected no files, written=%v skipped=%v", written, skipped)
	}
}

func TestWriteFilesSkipsExistingUnlessOverwrite(t *testing.T) {
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	if err := os.WriteFile("vercel.json", []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("DEPLOY.md", []byte("old-deploy"), 0644); err != nil {
		t.Fatal(err)
	}
	written, skipped, err := WriteFiles(ProviderVercel, "", func(string) bool { return false })
	if err != nil {
		t.Fatal(err)
	}
	if len(written) != 0 {
		t.Fatalf("expected no writes when overwrite denied, got %v", written)
	}
	if len(skipped) < 1 {
		t.Fatalf("expected skipped vercel.json, got skipped=%v", skipped)
	}
	data, _ := os.ReadFile(filepath.Join(dir, "vercel.json"))
	if string(data) != "old" {
		t.Fatalf("file should be unchanged")
	}

	written2, _, err := WriteFiles(ProviderVercel, "", func(string) bool { return true })
	if err != nil {
		t.Fatal(err)
	}
	if len(written2) < 1 {
		t.Fatalf("expected overwrite, written=%v", written2)
	}
}

func TestWriteFilesSkipsDEPLOYMDWhenExists(t *testing.T) {
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	if err := os.WriteFile("DEPLOY.md", []byte("keep"), 0644); err != nil {
		t.Fatal(err)
	}
	_, skipped, err := WriteFiles(ProviderVercel, "", func(string) bool { return false })
	if err != nil {
		t.Fatal(err)
	}
	if len(skipped) < 1 {
		t.Fatalf("expected DEPLOY.md skipped, got %v", skipped)
	}
	data, _ := os.ReadFile(filepath.Join(dir, "DEPLOY.md"))
	if string(data) != "keep" {
		t.Fatalf("DEPLOY.md should stay unchanged")
	}
}

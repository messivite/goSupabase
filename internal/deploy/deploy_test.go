package deploy

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNormalizeProvider(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"vercel", ProviderVercel},
		{"Vercel", ProviderVercel},
		{"1", ProviderVercel},
		{"fly", ProviderFly},
		{"", ProviderNone},
		{"nope", ProviderNone},
	}
	for _, tt := range tests {
		if got := NormalizeProvider(tt.in); got != tt.want {
			t.Fatalf("NormalizeProvider(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestWriteFilesVercel(t *testing.T) {
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	written, skipped, err := WriteFiles(ProviderVercel, "", func(string) bool { return true })
	if err != nil {
		t.Fatal(err)
	}
	if len(skipped) != 0 {
		t.Fatalf("skipped = %v", skipped)
	}
	if len(written) < 2 {
		t.Fatalf("expected vercel.json and DEPLOY.md, got %v", written)
	}
	for _, name := range []string{"vercel.json", "DEPLOY.md"} {
		p := filepath.Join(dir, name)
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("missing %s: %v", name, err)
		}
	}
}

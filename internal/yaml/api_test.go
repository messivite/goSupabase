package yamlcfg

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAndSave(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "api.yaml")

	cfg := &APIConfig{
		Version:  "1",
		BasePath: "/api",
		Endpoints: []Endpoint{
			{Method: "GET", Path: "/health", Handler: "Health"},
		},
	}
	if err := Save(path, cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.Version != "1" {
		t.Errorf("version = %q, want %q", loaded.Version, "1")
	}
	if len(loaded.Endpoints) != 1 {
		t.Fatalf("endpoints count = %d, want 1", len(loaded.Endpoints))
	}
	if loaded.Endpoints[0].Handler != "Health" {
		t.Errorf("handler = %q, want %q", loaded.Endpoints[0].Handler, "Health")
	}
}

func TestLoadFileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/api.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestAddEndpoint(t *testing.T) {
	cfg := &APIConfig{
		Endpoints: []Endpoint{
			{Method: "GET", Path: "/health", Handler: "Health"},
		},
	}

	err := cfg.AddEndpoint(Endpoint{Method: "POST", Path: "/tracks"})
	if err != nil {
		t.Fatalf("AddEndpoint: %v", err)
	}
	if len(cfg.Endpoints) != 2 {
		t.Fatalf("endpoints count = %d, want 2", len(cfg.Endpoints))
	}
	if cfg.Endpoints[1].Handler != "PostTracks" {
		t.Errorf("derived handler = %q, want %q", cfg.Endpoints[1].Handler, "PostTracks")
	}
}

func TestAddEndpointDuplicate(t *testing.T) {
	cfg := &APIConfig{
		Endpoints: []Endpoint{
			{Method: "GET", Path: "/health", Handler: "Health"},
		},
	}

	err := cfg.AddEndpoint(Endpoint{Method: "GET", Path: "/health", Handler: "Health"})
	if err == nil {
		t.Fatal("expected duplicate error")
	}
}

func TestAddEndpointDuplicateCaseInsensitive(t *testing.T) {
	cfg := &APIConfig{
		Endpoints: []Endpoint{
			{Method: "get", Path: "/health", Handler: "Health"},
		},
	}
	err := cfg.AddEndpoint(Endpoint{Method: "GET", Path: "/health"})
	if err == nil {
		t.Fatal("expected duplicate error for case-insensitive method match")
	}
}

func TestDeriveHandlerName(t *testing.T) {
	tests := []struct {
		method, path, want string
	}{
		{"POST", "/tracks", "PostTracks"},
		{"GET", "/tracks/:id", "GetTracksById"},
		{"DELETE", "/tracks/:id", "DeleteTracksById"},
		{"PATCH", "/tracks/:id", "PatchTracksById"},
		{"GET", "/health", "GetHealth"},
		{"GET", "/api/v1/users", "GetApiV1Users"},
	}
	for _, tt := range tests {
		got := DeriveHandlerName(tt.method, tt.path)
		if got != tt.want {
			t.Errorf("DeriveHandlerName(%q, %q) = %q, want %q", tt.method, tt.path, got, tt.want)
		}
	}
}

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"/tracks/:id", "/tracks/{id}"},
		{"/tracks", "/tracks"},
		{"/users/:userId/posts/:postId", "/users/{userId}/posts/{postId}"},
	}
	for _, tt := range tests {
		got := NormalizePath(tt.input)
		if got != tt.want {
			t.Errorf("NormalizePath(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestListEndpoints(t *testing.T) {
	cfg := &APIConfig{
		Endpoints: []Endpoint{
			{Method: "GET", Path: "/a"},
			{Method: "POST", Path: "/b"},
		},
	}
	eps := cfg.ListEndpoints()
	if len(eps) != 2 {
		t.Fatalf("ListEndpoints count = %d, want 2", len(eps))
	}
}

func TestSaveCreatesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.yaml")

	cfg := &APIConfig{Version: "1"}
	if err := Save(path, cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("file not created: %v", err)
	}
}

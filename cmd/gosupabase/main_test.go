package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
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

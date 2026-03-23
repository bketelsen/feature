package options

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetOptionsForPath(t *testing.T) {
	// Create a temp directory with a valid devcontainer-feature.json
	dir := t.TempDir()
	content := `{
		"id": "test-feature",
		"version": "1.0.0",
		"name": "Test Feature",
		"description": "A test feature",
		"containerEnv": {
			"FOO": "bar"
		}
	}`
	if err := os.WriteFile(filepath.Join(dir, "devcontainer-feature.json"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	opts, err := GetOptionsForPath(dir)
	if err != nil {
		t.Fatalf("GetOptionsForPath() error: %v", err)
	}
	if opts.ID != "test-feature" {
		t.Errorf("ID = %q, want %q", opts.ID, "test-feature")
	}
	if opts.Version != "1.0.0" {
		t.Errorf("Version = %q, want %q", opts.Version, "1.0.0")
	}
	if opts.ContainerEnv["FOO"] != "bar" {
		t.Errorf("ContainerEnv[FOO] = %q, want %q", opts.ContainerEnv["FOO"], "bar")
	}
}

func TestGetOptionsForPath_MissingFile(t *testing.T) {
	dir := t.TempDir()
	_, err := GetOptionsForPath(dir)
	if err == nil {
		t.Error("GetOptionsForPath() expected error for missing file, got nil")
	}
}

package scan

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectProjectRootPrefersMarker(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	project := filepath.Join(root, "demo")
	target := filepath.Join(project, "node_modules")

	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatalf("mkdir target: %v", err)
	}
	if err := os.WriteFile(filepath.Join(project, "package.json"), []byte("{}"), 0o644); err != nil {
		t.Fatalf("write marker: %v", err)
	}

	got := detectProjectRoot(target, root)
	if got != project {
		t.Fatalf("detectProjectRoot = %q, want %q", got, project)
	}
}

func TestScanFindsMatchingDirectory(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	project := filepath.Join(root, "api")
	target := filepath.Join(project, "node_modules")

	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatalf("mkdir target: %v", err)
	}
	if err := os.WriteFile(filepath.Join(project, "package.json"), []byte("{}"), 0o644); err != nil {
		t.Fatalf("write marker: %v", err)
	}
	if err := os.WriteFile(filepath.Join(target, "index.js"), []byte("hello"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	result, err := Scan(Options{Roots: []string{root}, Names: []string{"node_modules"}})
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}
	if len(result.Candidates) != 1 {
		t.Fatalf("Scan returned %d candidates, want 1", len(result.Candidates))
	}
	if result.Candidates[0].Project != project {
		t.Fatalf("candidate project = %q, want %q", result.Candidates[0].Project, project)
	}
}

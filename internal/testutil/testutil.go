// Package testutil provides shared helpers for building fixture trees.
package testutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// WriteTree creates files under root from a map of slash-separated relative
// paths to contents. Keys ending in "/" create (possibly empty) directories.
// Parent directories are created as needed.
func WriteTree(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for rel, content := range files {
		p := filepath.Join(root, filepath.FromSlash(strings.TrimSuffix(rel, "/")))
		if strings.HasSuffix(rel, "/") {
			if err := os.MkdirAll(p, 0o755); err != nil {
				t.Fatalf("creating dir %s: %v", p, err)
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatalf("creating parent of %s: %v", p, err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatalf("writing %s: %v", p, err)
		}
	}
}

// Chmod applies a mode to a path under root, registering cleanup to restore
// permissions so t.TempDir removal succeeds.
func Chmod(t *testing.T, root, rel string, mode os.FileMode) {
	t.Helper()
	p := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.Chmod(p, mode); err != nil {
		t.Fatalf("chmod %s: %v", p, err)
	}
	t.Cleanup(func() { _ = os.Chmod(p, 0o755) })
}

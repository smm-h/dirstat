package test

import (
	"flag"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/smm-h/dirstat/internal/testutil"
)

var update = flag.Bool("update", false, "rewrite golden files from current output")

var versionRe = regexp.MustCompile(`"dirstat_version": "[^"]*"`)

// goldenTree builds the fixed fixture tree used by the golden tests. It
// deliberately avoids symlinks and permission tricks so the JSON is
// identical on every platform.
func goldenTree(t *testing.T) string {
	root := t.TempDir()
	testutil.WriteTree(t, root, map[string]string{
		"a.go":       "package main\n\nfunc main() {}\n",
		"b.go":       "package b\n",
		"README.md":  "# Title\n\nBody text.\n",
		"data.bin":   "\x00\x01\x02\x03\x04",
		"Makefile":   "all:\n\techo hi\n",
		"sub/c.txt":  "one\ntwo\n",
		"sub/empty/": "",
	})
	return root
}

// normalize replaces the run-specific absolute root and binary version with
// stable placeholders.
func normalize(out, root string) string {
	out = versionRe.ReplaceAllString(out, `"dirstat_version": "VERSION"`)
	// The root only appears in the "root" field; paths in the document are
	// relative.
	rootJSON, _ := filepath.Abs(root)
	return regexp.MustCompile(regexp.QuoteMeta(rootJSON)).ReplaceAllString(out, "ROOT")
}

func checkGolden(t *testing.T, goldenName string, args ...string) {
	t.Helper()
	root := goldenTree(t)
	stdout, stderr, code := runDirstat(t, append([]string{"scan", root}, args...)...)
	if code != 0 {
		t.Fatalf("dirstat exited %d: %s", code, stderr)
	}
	got := normalize(stdout, root)

	goldenPath := filepath.Join("testdata", goldenName)
	if *update {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(goldenPath, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		t.Logf("updated %s", goldenPath)
		return
	}

	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("reading golden file (run with -update to create): %v", err)
	}
	if got != string(want) {
		t.Errorf("output differs from %s:\n--- got ---\n%s\n--- want ---\n%s", goldenPath, got, want)
	}
}

func TestGoldenJSONFull(t *testing.T) {
	checkGolden(t, "full.json", "--output", "json", "--list-no-ext")
}

func TestGoldenJSONSubset(t *testing.T) {
	checkGolden(t, "subset.json",
		"--output", "json",
		"--stats", "count", "--stats", "total-size", "--stats", "total-loc",
		"--type", "text",
		"--sort-by", "format", "--sort-order", "asc",
		"--method", "ext")
}

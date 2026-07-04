package scan

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
)

// ignoreMatcher matches scan-relative paths against gitignore patterns
// loaded from the enclosing git work tree.
type ignoreMatcher struct {
	m      gitignore.Matcher
	prefix []string // path components from the work tree root to the scan root
}

// newIgnoreMatcher builds a matcher for the work tree enclosing root, or
// returns nil when root is not inside a git work tree (then exclude/only
// behave as if no patterns exist) or when no patterns are found.
//
// Pattern sources, in ascending precedence (matching git): the global
// core.excludesFile, .git/info/exclude, and every .gitignore in the work
// tree (deeper files override shallower ones).
func newIgnoreMatcher(root string) *ignoreMatcher {
	wtRoot := findWorktreeRoot(root)
	if wtRoot == "" {
		return nil
	}

	var ps []gitignore.Pattern
	if global, err := gitignore.LoadGlobalPatterns(osfs.New("/")); err == nil {
		ps = append(ps, global...)
	}
	// ReadPatterns reads .git/info/exclude plus every .gitignore in the
	// tree, skipping .git and already-ignored directories.
	if tree, err := gitignore.ReadPatterns(osfs.New(wtRoot), nil); err == nil {
		ps = append(ps, tree...)
	}
	if len(ps) == 0 {
		return nil
	}

	var prefix []string
	if rel, err := filepath.Rel(wtRoot, root); err == nil && rel != "." {
		prefix = strings.Split(filepath.ToSlash(rel), "/")
	}
	return &ignoreMatcher{m: gitignore.NewMatcher(ps), prefix: prefix}
}

// findWorktreeRoot walks up from dir looking for a .git entry (directory or
// file). Returns "" when none is found.
func findWorktreeRoot(dir string) string {
	for {
		if _, err := os.Lstat(filepath.Join(dir, ".git")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

// Match reports whether the slash-separated path relative to the scan root
// is gitignored.
func (im *ignoreMatcher) Match(rel string, isDir bool) bool {
	parts := append(append([]string{}, im.prefix...), strings.Split(rel, "/")...)
	return im.m.Match(parts, isDir)
}

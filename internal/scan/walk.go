package scan

import (
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// walker performs the single-threaded directory traversal, collecting file
// entries for the classification pool and traversal-level summary stats.
type walker struct {
	opts  *Options
	ign   *ignoreMatcher // nil when Ignored == include or not in a git work tree
	excl  map[string]struct{}
	sum   Summary
	files []fileEntry
}

// absJoin joins the absolute root with a slash-separated relative path.
func absJoin(root, rel string) string {
	if rel == "" {
		return root
	}
	return filepath.Join(root, filepath.FromSlash(rel))
}

// walkRoot traverses the root directory. A read error on the root itself is
// returned; errors below the root are counted as unreadable.
func (w *walker) walkRoot() error {
	entries, err := os.ReadDir(w.opts.Root)
	if err != nil {
		return err
	}
	w.walkEntries("", 0, false, entries)
	return nil
}

// walkDir traverses one subdirectory (already accepted by the filters).
func (w *walker) walkDir(rel string, depth int, inIgnored bool) {
	entries, err := os.ReadDir(absJoin(w.opts.Root, rel))
	if err != nil {
		// Unreadable directory: count it and process whatever entries
		// os.ReadDir managed to return before failing.
		w.sum.Unreadable++
	}
	w.walkEntries(rel, depth, inIgnored, entries)
}

// walkEntries processes the (name-sorted) entries of the directory at rel,
// which sits at the given depth. inIgnored marks directories inside a
// gitignored directory: everything below one counts as ignored (relevant in
// "only" mode).
func (w *walker) walkEntries(rel string, depth int, inIgnored bool, entries []os.DirEntry) {
	for _, e := range entries {
		name := e.Name()

		// Symlinks are counted before any name-based filtering, so the
		// symlink count does not depend on filter flags (R19). They are
		// never followed.
		if e.Type()&fs.ModeSymlink != 0 {
			w.sum.Symlinks++
			continue
		}

		if _, excluded := w.excl[name]; excluded {
			continue
		}
		if w.opts.Hidden == HiddenExclude && strings.HasPrefix(name, ".") {
			continue
		}
		childRel := path.Join(rel, name)

		if e.IsDir() {
			childDepth := depth + 1
			if w.opts.Depth >= 0 && childDepth > w.opts.Depth {
				continue
			}
			childIgnored := inIgnored
			if !childIgnored && w.ign != nil && w.ign.Match(childRel, true) {
				childIgnored = true
			}
			if w.opts.Ignored == IgnoredExclude && childIgnored {
				continue
			}
			// In "only" mode non-ignored directories are still traversed
			// (an ignored directory may sit below them) but not counted.
			if w.opts.Ignored != IgnoredOnly || childIgnored {
				w.sum.Directories++
				if childDepth > w.sum.MaxDepth {
					w.sum.MaxDepth = childDepth
				}
			}
			w.walkDir(childRel, childDepth, childIgnored)
			continue
		}

		// Non-regular files (FIFOs, sockets, devices) are skipped silently.
		if e.Type() != 0 {
			continue
		}

		fileIgnored := inIgnored
		if !fileIgnored && w.ign != nil && w.ign.Match(childRel, false) {
			fileIgnored = true
		}
		if w.opts.Ignored == IgnoredExclude && fileIgnored {
			continue
		}
		if w.opts.Ignored == IgnoredOnly && !fileIgnored {
			continue
		}

		info, err := e.Info()
		if err != nil {
			w.sum.Unreadable++
			continue
		}
		w.files = append(w.files, fileEntry{
			rel:  childRel,
			name: name,
			size: info.Size(),
			exec: info.Mode().Perm()&0o111 != 0,
		})
	}
}

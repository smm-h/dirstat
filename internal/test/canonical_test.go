package test

import (
	"strings"
	"testing"

	"github.com/smm-h/dirstat/internal/testutil"
	"github.com/smm-h/stricttest/go/hygiene"
)

// --formats canonical merges alias formats into one group and resolves
// extensionless scripts by their shebang interpreter; --formats raw (the
// default) keeps every group name exactly as the extension or the sniffed
// MIME type produced it.

func TestCanonicalMergesAliasFormats(t *testing.T) {
	hygiene.Isolate(t, hygiene.Preserve(hygiene.GoPath, hygiene.GoModCache, hygiene.GoCache))
	root := t.TempDir()
	testutil.WriteTree(t, root, map[string]string{
		"a.mjs":   "export const a = 1;\n",
		"b.js":    "const b = 2;\n",
		"lib.h":   "#pragma once\n",
		"lib.c":   "int main(void) { return 0; }\n",
		"cls.hpp": "#pragma once\n",
	})

	raw := scanJSON(t, "scan", root, "--json", "--formats", "raw",
		"--sort-by", "format", "--sort-order", "asc")
	if got := strings.Join(groupFormats(t, raw), ","); got != "c,h,hpp,js,mjs" {
		t.Errorf("raw formats = %s, want c,h,hpp,js,mjs", got)
	}

	canon := scanJSON(t, "scan", root, "--json", "--formats", "canonical",
		"--sort-by", "format", "--sort-order", "asc")
	if got := strings.Join(groupFormats(t, canon), ","); got != "c,cpp,js" {
		t.Errorf("canonical formats = %s, want c,cpp,js", got)
	}
	if g := findGroup(t, canon, "js"); g["count"] != float64(2) {
		t.Errorf("canonical js count = %v, want 2 (a.mjs + b.js)", g["count"])
	}
	if g := findGroup(t, canon, "c"); g["count"] != float64(2) {
		t.Errorf("canonical c count = %v, want 2 (lib.h + lib.c)", g["count"])
	}
	if got := summaryField(t, canon, "unique_formats"); got != 3 {
		t.Errorf("canonical unique_formats = %d, want 3", got)
	}
}

func TestCanonicalResolvesShebangInterpreter(t *testing.T) {
	hygiene.Isolate(t, hygiene.Preserve(hygiene.GoPath, hygiene.GoModCache, hygiene.GoCache))
	root := t.TempDir()
	testutil.WriteTree(t, root, map[string]string{
		// The uv runner form sniffs as plain text, so only the shebang can
		// name this script's format.
		"tool": "#!/usr/bin/env -S uv run python\nprint('hi')\n",
	})

	raw := scanJSON(t, "scan", root, "--json", "--formats", "raw")
	if got := strings.Join(groupFormats(t, raw), ","); got != "text/plain" {
		t.Errorf("raw formats = %s, want text/plain", got)
	}

	canon := scanJSON(t, "scan", root, "--json", "--formats", "canonical")
	if got := strings.Join(groupFormats(t, canon), ","); got != "py" {
		t.Errorf("canonical formats = %s, want py", got)
	}
	if g := findGroup(t, canon, "py"); g["text"] != true {
		t.Errorf("canonical py group: text = %v, want true", g["text"])
	}
}

package classify

import (
	"io"
	"os"
	"path"
	"strings"
)

// shebangReadLimit is how many bytes of a file's start are read looking for a
// shebang line. Interpreter lines are short; anything longer is not one.
const shebangReadLimit = 512

// interpreterFormats maps a shebang interpreter's base name (version suffix
// already stripped) to the format name its scripts are grouped under in
// canonical mode. Lookup is case-sensitive: `Rscript` is spelled that way and
// nothing else is.
var interpreterFormats = map[string]string{
	"bash": "sh",
	"sh":   "sh",
	"zsh":  "sh",
	"dash": "sh",
	"ksh":  "sh",

	"node": "js",
	"deno": "js",
	"bun":  "js",

	"perl": "pl",
	"ruby": "rb",
	"php":  "php",
	"fish": "fish",
	"awk":  "awk",
	"gawk": "awk",
	"lua":  "lua",

	"Rscript": "r",
}

// readShebang returns the file's first line when it starts with "#!", and ""
// otherwise (including when the file cannot be read: the caller sniffs next
// and reports the read failure from there).
func readShebang(absPath string) string {
	f, err := os.Open(absPath)
	if err != nil {
		return ""
	}
	defer f.Close()

	buf := make([]byte, shebangReadLimit)
	n, err := f.Read(buf)
	if n == 0 || (err != nil && err != io.EOF) {
		return ""
	}
	head := string(buf[:n])
	if !strings.HasPrefix(head, "#!") {
		return ""
	}
	if i := strings.IndexAny(head, "\n\r"); i >= 0 {
		head = head[:i]
	}
	return head
}

// ShebangFormat maps a shebang line to the format name its scripts are grouped
// under in canonical mode. The second result is false when the line is not a
// shebang, names no interpreter, or names one with no unambiguous format.
//
// The interpreter is found by stripping an `env` wrapper (its flags, its
// `-S`/`--split-string` forms, and any NAME=value assignments), then unwrapping
// the `uv run X` and `uvx X` runner forms. A trailing version suffix on the
// interpreter's base name is dropped, so `python3.12` reads as `python`.
func ShebangFormat(line string) (string, bool) {
	if !strings.HasPrefix(line, "#!") {
		return "", false
	}
	tokens := strings.Fields(strings.TrimPrefix(line, "#!"))
	if len(tokens) == 0 {
		return "", false
	}

	if path.Base(tokens[0]) == "env" {
		tokens = stripEnvPrefix(tokens[1:])
	}
	tokens = unwrapRunner(tokens)
	if len(tokens) == 0 {
		return "", false
	}

	name := stripVersionSuffix(path.Base(tokens[0]))
	if strings.HasPrefix(name, "python") {
		return "py", true
	}
	format, ok := interpreterFormats[name]
	return format, ok
}

// stripEnvPrefix drops the options and NAME=value assignments that follow
// `env`, returning the tokens starting at the command it runs.
func stripEnvPrefix(tokens []string) []string {
	for len(tokens) > 0 {
		tok := tokens[0]
		switch {
		case tok == "-S" || tok == "--split-string":
			// The rest of the line is the command, already split by Fields.
			tokens = tokens[1:]
		case strings.HasPrefix(tok, "-S") && len(tok) > 2:
			// `-Scommand ...`: the command starts inside this token.
			return append([]string{tok[2:]}, tokens[1:]...)
		case strings.HasPrefix(tok, "--split-string="):
			return append([]string{strings.TrimPrefix(tok, "--split-string=")}, tokens[1:]...)
		case tok == "-u" || tok == "--unset":
			// Takes the variable name as a separate argument.
			if len(tokens) < 2 {
				return nil
			}
			tokens = tokens[2:]
		case strings.HasPrefix(tok, "-"):
			tokens = tokens[1:]
		case strings.Contains(tok, "="):
			tokens = tokens[1:]
		default:
			return tokens
		}
	}
	return nil
}

// unwrapRunner resolves the runner forms `uv run X` and `uvx X` to X. A runner
// whose next token is an option is left alone: what it would run cannot be
// read off the line without knowing the runner's own flags.
func unwrapRunner(tokens []string) []string {
	if len(tokens) == 0 {
		return tokens
	}
	switch path.Base(tokens[0]) {
	case "uv":
		if len(tokens) >= 3 && tokens[1] == "run" && !strings.HasPrefix(tokens[2], "-") {
			return tokens[2:]
		}
	case "uvx":
		if len(tokens) >= 2 && !strings.HasPrefix(tokens[1], "-") {
			return tokens[1:]
		}
	}
	return tokens
}

// stripVersionSuffix drops a trailing version from an interpreter name:
// `python3.12` -> `python`, `node20` -> `node`. A name that is all digits and
// dots is returned unchanged.
func stripVersionSuffix(name string) string {
	trimmed := strings.TrimRight(name, "0123456789.")
	if trimmed == "" {
		return name
	}
	return trimmed
}

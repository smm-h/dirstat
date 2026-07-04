package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/smm-h/dirstat/internal/config"
	"github.com/smm-h/dirstat/internal/jsonout"
	"github.com/smm-h/dirstat/internal/render"
	"github.com/smm-h/dirstat/internal/scan"
	"github.com/smm-h/strictcli/go/strictcli"
	"golang.org/x/term"
)

func registerScanCmd(app *strictcli.App) {
	app.Command("scan", "Summarize the files under a directory, grouped by format", handleScan,
		strictcli.WithFlags(
			strictcli.StringFlag("method",
				"grouping method: ext (extension only, no sniffing), type (content-sniff every file), hybrid (sniff extensionless files only)",
				strictcli.Choices("ext", "type", "hybrid"), strictcli.Default("hybrid")),
			strictcli.IntFlag("depth",
				"maximum directory depth below the root; -1 = unlimited; the root is depth 0",
				strictcli.Default(-1)),
			strictcli.StringFlag("exclude",
				"exact base name to skip, matching directories and files; repeatable",
				strictcli.Repeatable(), strictcli.Unique(true), strictcli.Default(nil)),
			strictcli.StringFlag("ignored",
				"gitignored-path handling: include (no matching at all), exclude (drop ignored paths), only (keep only ignored paths); when the target is not inside a git work tree, exclude and only behave as if no patterns exist",
				strictcli.Choices("include", "exclude", "only"), strictcli.Default("exclude")),
			strictcli.StringFlag("hidden",
				"dot-prefixed entry handling: include or exclude; the root directory itself is never treated as hidden",
				strictcli.Choices("include", "exclude"), strictcli.Default("include")),
			strictcli.StringFlag("stats",
				"stat to compute and show (repeatable; default: all): count, total-size, min-size, max-size, avg-size, total-loc, min-loc, max-loc, avg-loc",
				strictcli.Repeatable(), strictcli.Unique(true), strictcli.Default(nil)),
			strictcli.StringFlag("type",
				"filter groups by text/binary classification: text, binary, or both",
				strictcli.Choices("text", "binary", "both"), strictcli.Default("both")),
			strictcli.StringFlag("sort-by",
				"sort column, in precedence order when repeated: format, count, total-size, min-size, max-size, avg-size, total-loc, min-loc, max-loc, avg-loc",
				strictcli.Repeatable(), strictcli.Unique(true), strictcli.Default([]interface{}{"count"})),
			strictcli.StringFlag("sort-order",
				"sort direction, applied to all sort columns",
				strictcli.Choices("asc", "desc"), strictcli.Default("desc")),
			strictcli.IntFlag("top",
				"keep only the first N groups after sorting (table output only); -1 or 0 = all",
				strictcli.Default(-1)),
			strictcli.StringFlag("output",
				"output format: table (colored terminal tables) or json (stable machine-readable schema)",
				strictcli.Choices("table", "json"), strictcli.Default("table")),
			strictcli.StringFlag("show",
				"sections to render: summary, table, or both (table output only)",
				strictcli.Choices("summary", "table", "both"), strictcli.Default("both")),
			strictcli.BoolFlag("combined",
				"render one merged table with text and binary rows distinguished by color, instead of separate text/binary tables",
				strictcli.Default(true)),
			strictcli.StringFlag("singletons",
				"show one-file formats as-is, or collapse them into a single (singletons) row (table output only)",
				strictcli.Choices("show", "collapse"), strictcli.Default("show")),
			strictcli.BoolFlag("list-no-ext",
				"list the paths of extensionless files (a section in table output, a field in JSON output)",
				strictcli.Default(false)),
			strictcli.BoolFlag("legend",
				"show the text/binary color legend under the combined table (needs active colors)",
				strictcli.Default(true)),
			strictcli.BoolFlag("colors",
				"use ANSI colors; auto-disabled when stdout is not a TTY",
				strictcli.Default(true)),
			strictcli.BoolFlag("human",
				"human-readable sizes and thousands separators (table output only)",
				strictcli.Default(true)),
			strictcli.StringFlag("style",
				"table border character set: unicode or ascii",
				strictcli.Choices("unicode", "ascii"), strictcli.Default("unicode")),
		),
		strictcli.WithArgs(
			strictcli.NewArg("where", "directory to summarize (default: current directory)",
				strictcli.ArgRequired(false), strictcli.ArgDefault(".")),
		),
	)
}

// errorf prints a handler error line to stderr. The "error:" prefix is
// colored with the theme's error color when stderr is a TTY and --colors is
// true; stderr coloring is independent of --output mode (R30).
func errorf(colors bool, format string, args ...interface{}) {
	colored := colors && term.IsTerminal(int(os.Stderr.Fd()))
	fmt.Fprint(os.Stderr, render.FormatError(colored, config.LoadTheme(render.DetectThemeName()), format, args...))
}

// stringList extracts a repeatable string flag's values.
func stringList(kwargs map[string]interface{}, key string) []string {
	raw, _ := kwargs[key].([]interface{})
	out := make([]string, 0, len(raw))
	for _, v := range raw {
		out = append(out, v.(string))
	}
	return out
}

func handleScan(kwargs map[string]interface{}) int {
	colorsFlag := kwargs["colors"].(bool)

	where := kwargs["where"].(string)
	root, err := filepath.Abs(where)
	if err != nil {
		errorf(colorsFlag, "resolving path %q: %s", where, err)
		return ExitUsage
	}
	info, err := os.Stat(root)
	if err != nil {
		errorf(colorsFlag, "directory does not exist: %s", root)
		return ExitUsage
	}
	if !info.IsDir() {
		errorf(colorsFlag, "path is not a directory: %s", root)
		return ExitUsage
	}

	// Validate --stats (empty = all).
	validStats := make(map[string]bool, len(scan.StatNames))
	for _, s := range scan.StatNames {
		validStats[s] = true
	}
	statsRaw := stringList(kwargs, "stats")
	selected := make(jsonout.Selection)
	if len(statsRaw) == 0 {
		for _, s := range scan.StatNames {
			selected[s] = true
		}
	} else {
		for _, s := range statsRaw {
			if !validStats[s] {
				errorf(colorsFlag, "invalid stat %q; valid: %v", s, scan.StatNames)
				return ExitUsage
			}
			selected[s] = true
		}
	}
	// Selected stats in canonical column order.
	var statsOrdered []string
	for _, s := range scan.StatNames {
		if selected[s] {
			statsOrdered = append(statsOrdered, s)
		}
	}

	// Validate --sort-by.
	sortBy := stringList(kwargs, "sort_by")
	if len(sortBy) == 0 {
		sortBy = []string{"count"}
	}
	for _, s := range sortBy {
		if s != "format" && !validStats[s] {
			errorf(colorsFlag, "invalid sort column %q; valid: format, %v", s, scan.StatNames)
			return ExitUsage
		}
	}
	sortDesc := kwargs["sort_order"].(string) == "desc"

	// LOC is expensive: read files only when a LOC stat is shown or sorted
	// by. Sorting by a LOC stat forces the computation even when the stat
	// is not selected for display.
	needLOC := false
	for s := range selected {
		if scan.IsLOCStat(s) {
			needLOC = true
		}
	}
	for _, s := range sortBy {
		if scan.IsLOCStat(s) {
			needLOC = true
		}
	}

	listNoExt := kwargs["list_no_ext"].(bool)

	res, err := scan.Scan(scan.Options{
		Root:       root,
		Method:     kwargs["method"].(string),
		Depth:      kwargs["depth"].(int),
		Exclude:    stringList(kwargs, "exclude"),
		Ignored:    kwargs["ignored"].(string),
		Hidden:     kwargs["hidden"].(string),
		TypeFilter: kwargs["type"].(string),
		NeedLOC:    needLOC,
		ListNoExt:  listNoExt,
		TextExts:   config.TextExtensions(),
		TextMimes:  config.TextMimetypes(),
	})
	if err != nil {
		errorf(colorsFlag, "%s", err)
		return ExitGeneral
	}

	if kwargs["output"].(string) == "json" {
		groups := make([]scan.Group, len(res.Groups))
		copy(groups, res.Groups)
		scan.SortGroups(groups, sortBy, sortDesc)
		if err := jsonout.Write(os.Stdout, version, res, groups, selected, listNoExt); err != nil {
			errorf(colorsFlag, "writing JSON: %s", err)
			return ExitGeneral
		}
		return ExitSuccess
	}

	// Colors apply only when requested AND stdout is a TTY; without active
	// colors the width is pinned to 80 so --no-colors output is
	// byte-identical to non-TTY output (R30).
	colorsActive := colorsFlag && term.IsTerminal(int(os.Stdout.Fd()))
	width := 80
	if colorsActive {
		if tw, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && tw > 0 {
			width = tw
		}
	}

	render.Render(os.Stdout, res, render.Options{
		Show:       kwargs["show"].(string),
		Combined:   kwargs["combined"].(bool),
		Collapse:   kwargs["singletons"].(string) == "collapse",
		SortBy:     sortBy,
		SortDesc:   sortDesc,
		Top:        kwargs["top"].(int),
		ListNoExt:  listNoExt,
		Legend:     kwargs["legend"].(bool),
		Colors:     colorsActive,
		Human:      kwargs["human"].(bool),
		Style:      kwargs["style"].(string),
		TypeFilter: kwargs["type"].(string),
		Stats:      statsOrdered,
		Width:      width,
		Theme:      config.LoadTheme(render.DetectThemeName()),
	})
	return ExitSuccess
}

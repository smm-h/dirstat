package render

import (
	"strings"
	"testing"

	"github.com/smm-h/dirstat/internal/config"
	"github.com/smm-h/dirstat/internal/scan"
)

func testResult() *scan.Result {
	return &scan.Result{
		Root:   "/tmp/x",
		Method: "ext",
		Summary: scan.Summary{
			Directories: 2, Files: 3, MaxDepth: 1, UniqueFormats: 2,
		},
		Groups: []scan.Group{
			{Format: "go", Text: true, Count: 2, TotalSize: 200, MinSize: 50, MaxSize: 150, AvgSize: 100,
				HasLOC: true, TotalLOC: 20, MinLOC: 5, MaxLOC: 15, AvgLOC: 10},
			{Format: "png", Text: false, Count: 1, TotalSize: 4096, MinSize: 4096, MaxSize: 4096, AvgSize: 4096},
		},
	}
}

func testOpts() Options {
	return Options{
		Show:       "both",
		Combined:   true,
		SortBy:     []string{"count"},
		SortDesc:   true,
		Top:        -1,
		Legend:     true,
		Colors:     false,
		Human:      true,
		Style:      "unicode",
		TypeFilter: scan.TypeBoth,
		Stats:      scan.StatNames,
		Width:      200,
		Theme:      config.LoadTheme("dark"),
	}
}

func renderToString(res *scan.Result, opts Options) string {
	var sb strings.Builder
	Render(&sb, res, opts)
	return sb.String()
}

func TestRenderSections(t *testing.T) {
	out := renderToString(testResult(), testOpts())
	for _, want := range []string{
		"Summary", "Directories: 2", "Files: 3", "Unreadable (skipped): 0",
		"Files by Format", "Format", "Count", "Total Size", "Avg LOC",
		"go", "png", "4.0KB", "┌", "│", "└",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
	// Binary rows get "-" for LOC columns.
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "png") && !strings.Contains(line, "-") {
			t.Errorf("png row should contain '-' LOC cells: %q", line)
		}
	}
	// No ANSI codes when colors are off.
	if strings.Contains(out, "\033[") {
		t.Error("output contains ANSI codes with colors disabled")
	}
}

func TestRenderShowFilter(t *testing.T) {
	opts := testOpts()
	opts.Show = "summary"
	out := renderToString(testResult(), opts)
	if !strings.Contains(out, "Summary") || strings.Contains(out, "Files by Format") {
		t.Errorf("show=summary rendered wrong sections:\n%s", out)
	}

	opts.Show = "table"
	out = renderToString(testResult(), opts)
	if strings.Contains(out, "Summary") || !strings.Contains(out, "Files by Format") {
		t.Errorf("show=table rendered wrong sections:\n%s", out)
	}
}

func TestRenderSplitMode(t *testing.T) {
	opts := testOpts()
	opts.Combined = false
	out := renderToString(testResult(), opts)
	if !strings.Contains(out, "Text Files by Format") || !strings.Contains(out, "Binary Files by Format") {
		t.Errorf("split mode missing table headings:\n%s", out)
	}
	// The binary table must omit LOC columns entirely: no "-" cells.
	binPart := out[strings.Index(out, "Binary Files by Format"):]
	if strings.Contains(binPart, "Total LOC") {
		t.Errorf("binary table must omit LOC columns:\n%s", binPart)
	}
}

func TestRenderAsciiStyle(t *testing.T) {
	opts := testOpts()
	opts.Style = "ascii"
	out := renderToString(testResult(), opts)
	if strings.ContainsAny(out, "┌┬┐├┼┤└┴┘│─") {
		t.Errorf("ascii style must not contain box-drawing characters:\n%s", out)
	}
	if !strings.Contains(out, "+--") || !strings.Contains(out, "| ") {
		t.Errorf("ascii style should use + - |:\n%s", out)
	}
}

func TestRenderEmptyResult(t *testing.T) {
	res := &scan.Result{Root: "/tmp/x", Method: "ext"}
	out := renderToString(res, testOpts())
	if !strings.Contains(out, "No files found.") {
		t.Errorf("empty result should print 'No files found.':\n%s", out)
	}
	if strings.Contains(out, "│") {
		t.Errorf("empty result should not render a table:\n%s", out)
	}
}

func TestRenderColors(t *testing.T) {
	opts := testOpts()
	opts.Colors = true
	out := renderToString(testResult(), opts)
	if !strings.Contains(out, "\033[38;5;114m") {
		t.Error("expected dark theme text color in output")
	}
	if !strings.Contains(out, "■") || !strings.Contains(out, "Text files") {
		t.Error("expected legend when colors active, combined, type both")
	}

	// Legend disappears when colors are off (byte-identical rule).
	opts.Colors = false
	out = renderToString(testResult(), opts)
	if strings.Contains(out, "■") {
		t.Error("legend must not render without colors")
	}
}

func TestRenderLegendConditions(t *testing.T) {
	base := testOpts()
	base.Colors = true

	noLegend := base
	noLegend.Legend = false
	if strings.Contains(renderToString(testResult(), noLegend), "■") {
		t.Error("legend rendered despite --no-legend")
	}

	split := base
	split.Combined = false
	if strings.Contains(renderToString(testResult(), split), "■") {
		t.Error("legend rendered in split mode")
	}

	textOnly := base
	textOnly.TypeFilter = scan.TypeText
	if strings.Contains(renderToString(testResult(), textOnly), "■") {
		t.Error("legend rendered with --type text")
	}
}

func TestRenderWidthShrinkAndTruncate(t *testing.T) {
	res := testResult()
	res.Groups[0].Format = "a-very-long-format-name-that-overflows-the-table"
	opts := testOpts()
	opts.Width = 70
	opts.Stats = []string{"count", "total-size", "min-size", "max-size", "avg-size"}
	out := renderToString(res, opts)
	if !strings.Contains(out, "..") {
		t.Errorf("expected truncated format cell with '..' suffix:\n%s", out)
	}
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "│") && len([]rune(line)) > 70 {
			t.Errorf("table line exceeds width %d: %q", opts.Width, line)
		}
	}

	// Below the floor of 10 the Format column stops shrinking, so the
	// table may legitimately overflow very narrow terminals.
	opts.Width = 20
	out = renderToString(res, opts)
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "..") && !strings.Contains(line, "a-very-l..") {
			t.Errorf("format column should be exactly 10 wide at the floor: %q", line)
		}
	}
}

func TestRenderTopAndCollapse(t *testing.T) {
	res := testResult()
	res.Groups = append(res.Groups, scan.Group{Format: "md", Text: true, Count: 1,
		TotalSize: 10, MinSize: 10, MaxSize: 10, AvgSize: 10, HasLOC: true,
		TotalLOC: 3, MinLOC: 3, MaxLOC: 3, AvgLOC: 3})

	opts := testOpts()
	opts.Collapse = true
	out := renderToString(res, opts)
	if !strings.Contains(out, "(singletons)") {
		t.Errorf("expected (singletons) row:\n%s", out)
	}
	if strings.Contains(out, " md ") {
		t.Errorf("md singleton should have been collapsed:\n%s", out)
	}

	opts = testOpts()
	opts.Top = 1
	out = renderToString(res, opts)
	if !strings.Contains(out, "go") || strings.Contains(out, "png") {
		t.Errorf("--top 1 should keep only the go group:\n%s", out)
	}
}

func TestRenderNoExtList(t *testing.T) {
	res := testResult()
	res.NoExtFiles = []string{"Makefile", "sub/LICENSE"}
	opts := testOpts()
	opts.ListNoExt = true
	out := renderToString(res, opts)
	// The summary also contains a "Files without extension" stat line, so
	// look for the last occurrence (the list heading).
	idx := strings.LastIndex(out, "Files without extension")
	if idx < 0 {
		t.Fatalf("missing extensionless list heading:\n%s", out)
	}
	if !strings.Contains(out[idx:], "Makefile") || !strings.Contains(out[idx:], "sub/LICENSE") {
		t.Errorf("extensionless list missing entries:\n%s", out)
	}
	if idx < strings.Index(out, "Files by Format") {
		t.Error("extensionless list must come after the table section")
	}
}

func TestDetectThemeName(t *testing.T) {
	tests := []struct {
		env  string
		want string
	}{
		{"", "dark"},
		{"15;0", "dark"},
		{"0;15", "light"},
		{"0;7", "light"},
		{"default;default", "dark"},
		{"12;8", "dark"},
		{"0;10", "light"},
	}
	for _, tc := range tests {
		t.Setenv("COLORFGBG", tc.env)
		if got := DetectThemeName(); got != tc.want {
			t.Errorf("COLORFGBG=%q: got %s, want %s", tc.env, got, tc.want)
		}
	}
}

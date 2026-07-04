package jsonout

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/smm-h/dirstat/internal/scan"
)

func fullSelection() Selection {
	sel := Selection{}
	for _, s := range scan.StatNames {
		sel[s] = true
	}
	return sel
}

func testResult() *scan.Result {
	return &scan.Result{
		Root:   "/data/project",
		Method: "hybrid",
		Summary: scan.Summary{
			Directories: 3, Files: 5, FilesWithoutExtension: 1, MaxDepth: 2,
			Symlinks: 1, Executables: 2, UniqueFormats: 2, FilesSniffed: 1, Unreadable: 0,
		},
		Groups: []scan.Group{
			{Format: "go", Text: true, Count: 4, TotalSize: 400, MinSize: 50, MaxSize: 200, AvgSize: 100,
				HasLOC: true, TotalLOC: 40, MinLOC: 2, MaxLOC: 20, AvgLOC: 10},
			{Format: "png", Text: false, Count: 1, TotalSize: 9000, MinSize: 9000, MaxSize: 9000, AvgSize: 9000},
		},
		NoExtFiles: []string{"Makefile"},
	}
}

func writeToString(t *testing.T, res *scan.Result, sel Selection, listNoExt bool) string {
	t.Helper()
	var sb strings.Builder
	if err := Write(&sb, "1.2.3", res, res.Groups, sel, listNoExt); err != nil {
		t.Fatal(err)
	}
	return sb.String()
}

func TestWriteFullSchema(t *testing.T) {
	res := testResult()
	out := writeToString(t, res, fullSelection(), true)

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, out)
	}
	if parsed["dirstat_version"] != "1.2.3" || parsed["root"] != "/data/project" || parsed["method"] != "hybrid" {
		t.Errorf("bad top-level fields: %v", parsed)
	}
	summary := parsed["summary"].(map[string]interface{})
	for key, want := range map[string]float64{
		"directories": 3, "files": 5, "files_without_extension": 1, "max_depth": 2,
		"symlinks": 1, "executables": 2, "unique_formats": 2, "files_sniffed": 1, "unreadable": 0,
	} {
		if summary[key] != want {
			t.Errorf("summary[%s] = %v, want %v", key, summary[key], want)
		}
	}

	groups := parsed["groups"].([]interface{})
	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}
	goGroup := groups[0].(map[string]interface{})
	if goGroup["format"] != "go" || goGroup["text"] != true || goGroup["total_loc"] != float64(40) {
		t.Errorf("bad go group: %v", goGroup)
	}
	pngGroup := groups[1].(map[string]interface{})
	if pngGroup["total_loc"] != nil {
		t.Errorf("binary group total_loc should be null, got %v", pngGroup["total_loc"])
	}
	if v, present := pngGroup["min_loc"]; !present || v != nil {
		t.Errorf("binary group min_loc should be present and null, got %v (present=%v)", v, present)
	}

	noExt := parsed["no_extension_files"].([]interface{})
	if len(noExt) != 1 || noExt[0] != "Makefile" {
		t.Errorf("bad no_extension_files: %v", noExt)
	}

	// No ANSI codes ever.
	if strings.Contains(out, "\033") {
		t.Error("JSON output contains ANSI escape")
	}
	if !strings.HasSuffix(out, "\n") {
		t.Error("JSON output must end with a newline")
	}
}

func TestWriteFieldOrder(t *testing.T) {
	out := writeToString(t, testResult(), fullSelection(), false)
	wantOrder := []string{
		`"dirstat_version"`, `"root"`, `"method"`, `"summary"`,
		`"directories"`, `"files"`, `"files_without_extension"`, `"max_depth"`,
		`"symlinks"`, `"executables"`, `"unique_formats"`, `"files_sniffed"`, `"unreadable"`,
		`"groups"`, `"format"`, `"text"`, `"count"`,
		`"total_size"`, `"min_size"`, `"max_size"`, `"avg_size"`,
		`"total_loc"`, `"min_loc"`, `"max_loc"`, `"avg_loc"`,
	}
	pos := -1
	for _, key := range wantOrder {
		i := strings.Index(out, key)
		if i < 0 {
			t.Fatalf("missing key %s:\n%s", key, out)
		}
		if i < pos {
			t.Errorf("key %s out of order", key)
		}
		pos = i
	}
}

func TestWriteStatsSelection(t *testing.T) {
	sel := Selection{"count": true, "total-size": true}
	out := writeToString(t, testResult(), sel, false)
	for _, absent := range []string{`"min_size"`, `"avg_size"`, `"total_loc"`, `"avg_loc"`} {
		if strings.Contains(out, absent) {
			t.Errorf("unselected stat %s present in output:\n%s", absent, out)
		}
	}
	for _, present := range []string{`"count"`, `"total_size"`, `"format"`, `"text"`} {
		if !strings.Contains(out, present) {
			t.Errorf("expected %s in output:\n%s", present, out)
		}
	}
}

func TestWriteNoExtOmittedAndEmpty(t *testing.T) {
	res := testResult()
	out := writeToString(t, res, fullSelection(), false)
	if strings.Contains(out, "no_extension_files") {
		t.Error("no_extension_files must be omitted without --list-no-ext")
	}

	res.NoExtFiles = nil
	out = writeToString(t, res, fullSelection(), true)
	if !strings.Contains(out, `"no_extension_files": []`) {
		t.Errorf("expected empty array for no_extension_files:\n%s", out)
	}
}

func TestWriteEmptyGroups(t *testing.T) {
	res := &scan.Result{Root: "/x", Method: "ext"}
	out := writeToString(t, res, fullSelection(), false)
	if !strings.Contains(out, `"groups": []`) {
		t.Errorf("expected empty groups array:\n%s", out)
	}
}

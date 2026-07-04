// Package jsonout emits the machine-readable JSON output. The schema is a
// consumer contract (R34): field names are stable, values are raw integers,
// no ANSI codes, and evolution is additive-only.
package jsonout

import (
	"bytes"
	"encoding/json"
	"io"
	"strconv"

	"github.com/smm-h/dirstat/internal/scan"
)

// Selection reports which stats were selected via --stats. Unselected stats
// are omitted from group objects.
type Selection map[string]bool

type summaryJSON struct {
	Directories           int `json:"directories"`
	Files                 int `json:"files"`
	FilesWithoutExtension int `json:"files_without_extension"`
	MaxDepth              int `json:"max_depth"`
	Symlinks              int `json:"symlinks"`
	Executables           int `json:"executables"`
	UniqueFormats         int `json:"unique_formats"`
	FilesSniffed          int `json:"files_sniffed"`
	Unreadable            int `json:"unreadable"`
}

type groupJSON struct {
	g   scan.Group
	sel Selection
}

// MarshalJSON writes the group fields in the documented, stable order.
// Unselected stats are omitted; LOC fields are null for binary groups.
func (gj groupJSON) MarshalJSON() ([]byte, error) {
	var b bytes.Buffer
	b.WriteByte('{')

	writeKey := func(key string) {
		if b.Len() > 1 {
			b.WriteByte(',')
		}
		b.WriteByte('"')
		b.WriteString(key)
		b.WriteString(`":`)
	}
	writeInt := func(key string, v int64) {
		writeKey(key)
		b.WriteString(strconv.FormatInt(v, 10))
	}
	writeLOC := func(key string, v int64) {
		writeKey(key)
		if gj.g.HasLOC {
			b.WriteString(strconv.FormatInt(v, 10))
		} else {
			b.WriteString("null")
		}
	}

	name, err := json.Marshal(gj.g.Format)
	if err != nil {
		return nil, err
	}
	writeKey("format")
	b.Write(name)
	writeKey("text")
	b.WriteString(strconv.FormatBool(gj.g.Text))

	if gj.sel["count"] {
		writeInt("count", int64(gj.g.Count))
	}
	if gj.sel["total-size"] {
		writeInt("total_size", gj.g.TotalSize)
	}
	if gj.sel["min-size"] {
		writeInt("min_size", gj.g.MinSize)
	}
	if gj.sel["max-size"] {
		writeInt("max_size", gj.g.MaxSize)
	}
	if gj.sel["avg-size"] {
		writeInt("avg_size", gj.g.AvgSize)
	}
	if gj.sel["total-loc"] {
		writeLOC("total_loc", gj.g.TotalLOC)
	}
	if gj.sel["min-loc"] {
		writeLOC("min_loc", gj.g.MinLOC)
	}
	if gj.sel["max-loc"] {
		writeLOC("max_loc", gj.g.MaxLOC)
	}
	if gj.sel["avg-loc"] {
		writeLOC("avg_loc", gj.g.AvgLOC)
	}

	b.WriteByte('}')
	return b.Bytes(), nil
}

type doc struct {
	Version string      `json:"dirstat_version"`
	Root    string      `json:"root"`
	Method  string      `json:"method"`
	Config  string      `json:"config,omitempty"`
	Summary summaryJSON `json:"summary"`
	Groups  []groupJSON `json:"groups"`
	NoExt   *[]string   `json:"no_extension_files,omitempty"`
}

// Write emits the JSON document for a scan result. groups must already be
// sorted per --sort-by/--sort-order. listNoExt controls the presence of the
// no_extension_files field. configPath is the absolute path of the scan
// config file in effect; when empty the additive "config" field is omitted
// entirely (R46).
func Write(w io.Writer, version, configPath string, res *scan.Result, groups []scan.Group, sel Selection, listNoExt bool) error {
	s := res.Summary
	d := doc{
		Version: version,
		Root:    res.Root,
		Method:  res.Method,
		Config:  configPath,
		Summary: summaryJSON{
			Directories:           s.Directories,
			Files:                 s.Files,
			FilesWithoutExtension: s.FilesWithoutExtension,
			MaxDepth:              s.MaxDepth,
			Symlinks:              s.Symlinks,
			Executables:           s.Executables,
			UniqueFormats:         s.UniqueFormats,
			FilesSniffed:          s.FilesSniffed,
			Unreadable:            s.Unreadable,
		},
		Groups: make([]groupJSON, len(groups)),
	}
	for i, g := range groups {
		d.Groups[i] = groupJSON{g: g, sel: sel}
	}
	if listNoExt {
		files := res.NoExtFiles
		if files == nil {
			files = []string{}
		}
		d.NoExt = &files
	}

	out, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return err
	}
	out = append(out, '\n')
	_, err = w.Write(out)
	return err
}

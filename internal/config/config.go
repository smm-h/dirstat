// Package config provides the embedded text-classification lists and color
// themes. Everything is compiled into the binary via go:embed: there are no
// runtime config files and no lookup in cwd, HOME, or XDG directories.
// Changing the lists means a rebuild.
package config

import (
	"embed"
	"encoding/json"
	"strings"
)

//go:embed data/text_extensions.txt data/text_mimetypes.txt data/colors_dark.json data/colors_light.json
var dataFS embed.FS

// parseList extracts whitespace-separated, lowercased tokens from an embedded
// list file, skipping blank lines and #-comment lines.
func parseList(raw string) map[string]struct{} {
	set := make(map[string]struct{})
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		for _, tok := range strings.Fields(line) {
			tok = strings.TrimPrefix(strings.ToLower(tok), ".")
			if tok != "" {
				set[tok] = struct{}{}
			}
		}
	}
	return set
}

func mustRead(name string) string {
	b, err := dataFS.ReadFile("data/" + name)
	if err != nil {
		panic("dirstat: embedded data file missing: " + name)
	}
	return string(b)
}

// TextExtensions returns the set of lowercased, dot-stripped file extensions
// that classify a file as text.
func TextExtensions() map[string]struct{} {
	return parseList(mustRead("text_extensions.txt"))
}

// TextMimetypes returns the set of lowercased MIME types that classify a file
// as text even though they do not start with "text/".
func TextMimetypes() map[string]struct{} {
	return parseList(mustRead("text_mimetypes.txt"))
}

// Theme holds the ANSI 256-color codes for one color theme. Each value is a
// 256-color palette index as a decimal string, or "" for no color.
type Theme struct {
	TextFG    string `json:"text_fg"`
	TextBG    string `json:"text_bg"`
	BinaryFG  string `json:"binary_fg"`
	BinaryBG  string `json:"binary_bg"`
	HeaderFG  string `json:"header_fg"`
	HeaderBG  string `json:"header_bg"`
	StatLabel string `json:"stat_label"`
	StatValue string `json:"stat_value"`
	Error     string `json:"error"`
	Border    string `json:"border"`
}

// LoadTheme returns the embedded theme for name, which must be "dark" or
// "light".
func LoadTheme(name string) Theme {
	if name != "dark" && name != "light" {
		panic("dirstat: unknown theme: " + name)
	}
	var t Theme
	if err := json.Unmarshal([]byte(mustRead("colors_"+name+".json")), &t); err != nil {
		panic("dirstat: embedded theme " + name + " is invalid JSON: " + err.Error())
	}
	return t
}

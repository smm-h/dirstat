package config

import (
	"testing"

	"github.com/smm-h/stricttest/go/hygiene"
)

func TestTextExtensions(t *testing.T) {
	hygiene.Isolate(t, hygiene.Preserve(hygiene.GoPath, hygiene.GoModCache, hygiene.GoCache))
	exts := TextExtensions()
	if len(exts) < 90 {
		t.Fatalf("expected at least 90 text extensions, got %d", len(exts))
	}
	for _, tc := range []string{"svg", "go", "py", "md", "json", "sh", "makefile", "gitignore"} {
		if _, ok := exts[tc]; !ok {
			t.Errorf("expected extension %q in text extensions list", tc)
		}
	}
	for _, tc := range []string{"png", "exe", "", "#"} {
		if _, ok := exts[tc]; ok {
			t.Errorf("did not expect %q in text extensions list", tc)
		}
	}
	// Formats reported missing from the list: Godot's text formats, Go's
	// module files, and a batch of web/shader/schema languages.
	for _, tc := range []string{
		"gd", "tscn", "tres", "gdshader", "godot", "uid",
		"mod", "sum",
		"svx", "astro", "prisma", "odin",
		"wgsl", "glsl", "hlsl", "frag", "vert",
	} {
		if _, ok := exts[tc]; !ok {
			t.Errorf("expected extension %q in text extensions list", tc)
		}
	}
	// `lock` is deliberately absent: uv and Poetry lockfiles are TOML text
	// while other tools write binary .lock files, so hybrid mode must sniff
	// each one instead of trusting a static entry (R13).
	if _, ok := exts["lock"]; ok {
		t.Error("lock must stay out of the text extensions list so it is sniffed per file")
	}
}

func TestTextMimetypes(t *testing.T) {
	hygiene.Isolate(t, hygiene.Preserve(hygiene.GoPath, hygiene.GoModCache, hygiene.GoCache))
	mimes := TextMimetypes()
	if len(mimes) < 40 {
		t.Fatalf("expected at least 40 text mimetypes, got %d", len(mimes))
	}
	for _, tc := range []string{"image/svg+xml", "application/json", "application/x-empty", "inode/x-empty"} {
		if _, ok := mimes[tc]; !ok {
			t.Errorf("expected mimetype %q in text mimetypes list", tc)
		}
	}
	if _, ok := mimes["image/png"]; ok {
		t.Error("did not expect image/png in text mimetypes list")
	}
}

func TestLoadTheme(t *testing.T) {
	hygiene.Isolate(t, hygiene.Preserve(hygiene.GoPath, hygiene.GoModCache, hygiene.GoCache))
	tests := []struct {
		name   string
		textFG string
	}{
		{"dark", "114"},
		{"light", "28"},
	}
	for _, tc := range tests {
		th := LoadTheme(tc.name)
		if th.TextFG != tc.textFG {
			t.Errorf("theme %s: TextFG = %q, want %q", tc.name, th.TextFG, tc.textFG)
		}
		if th.Border == "" || th.HeaderFG == "" || th.StatLabel == "" || th.StatValue == "" || th.Error == "" {
			t.Errorf("theme %s: unexpected empty required color: %+v", tc.name, th)
		}
	}
}

func TestLoadThemeUnknownPanics(t *testing.T) {
	hygiene.Isolate(t, hygiene.Preserve(hygiene.GoPath, hygiene.GoModCache, hygiene.GoCache))
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for unknown theme name")
		}
	}()
	LoadTheme("solarized")
}

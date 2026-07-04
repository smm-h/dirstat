// Package classify determines a file's group name (format) and its
// text/binary classification, according to the selected grouping method.
package classify

import (
	"strings"

	"github.com/gabriel-vasile/mimetype"
)

// Grouping method names.
const (
	MethodExt    = "ext"
	MethodType   = "type"
	MethodHybrid = "hybrid"
)

// Group names for files that could not be assigned a real format.
const (
	GroupNoExtension = "(no extension)"
	GroupUnknown     = "(unknown)"
)

// Class is the classification outcome for a single file.
type Class struct {
	Group   string // group (format) name
	Text    bool   // strict text/binary classification
	NoExt   bool   // the file name has no extension
	Sniffed bool   // a content-sniff operation was performed
}

// Classifier classifies files per a fixed method and embedded lists.
type Classifier struct {
	method    string
	textExts  map[string]struct{}
	textMimes map[string]struct{}
}

// New returns a Classifier for the given method ("ext", "type", or "hybrid").
func New(method string, textExts, textMimes map[string]struct{}) *Classifier {
	return &Classifier{method: method, textExts: textExts, textMimes: textMimes}
}

// NormalizeExt returns the normalized extension of a file name: the suffix
// after the last dot, lowercased, without the dot. Following Python's
// Path.suffix semantics: "archive.tar.gz" -> "gz"; dotfiles like ".bashrc",
// bare names like "Makefile", and trailing-dot names have no extension.
func NormalizeExt(name string) string {
	i := strings.LastIndexByte(name, '.')
	if i <= 0 || i == len(name)-1 {
		return ""
	}
	return strings.ToLower(name[i+1:])
}

// ExtIsText reports whether a normalized extension classifies as text.
func (c *Classifier) ExtIsText(ext string) bool {
	_, ok := c.textExts[ext]
	return ok
}

// MimeIsText reports whether a MIME type classifies as text: it starts with
// "text/" or is in the embedded text-mimetypes list.
func (c *Classifier) MimeIsText(mime string) bool {
	if strings.HasPrefix(mime, "text/") {
		return true
	}
	_, ok := c.textMimes[mime]
	return ok
}

// sniff content-sniffs the file and returns the bare MIME type (parameters
// like "; charset=utf-8" stripped, lowercased). A non-nil error means the
// file could not be read.
func sniff(absPath string) (string, error) {
	m, err := mimetype.DetectFile(absPath)
	if err != nil {
		return "", err
	}
	s := m.String()
	if i := strings.IndexByte(s, ';'); i >= 0 {
		s = s[:i]
	}
	return strings.ToLower(strings.TrimSpace(s)), nil
}

// File classifies the file at absPath whose base name is name. A non-nil
// error means the file could not be read during sniffing; the caller must
// treat it as unreadable. Class.Sniffed is valid even when err is non-nil.
func (c *Classifier) File(absPath, name string) (Class, error) {
	ext := NormalizeExt(name)
	cls := Class{NoExt: ext == ""}

	byExt := func() {
		if ext == "" {
			cls.Group = GroupNoExtension
		} else {
			cls.Group = ext
		}
		cls.Text = c.ExtIsText(ext)
	}

	switch c.method {
	case MethodExt:
		byExt()
	case MethodType:
		cls.Sniffed = true
		mime, err := sniff(absPath)
		if err != nil {
			return cls, err
		}
		if mime == "" {
			cls.Group = GroupUnknown
			cls.Text = false
		} else {
			cls.Group = mime
			cls.Text = c.MimeIsText(mime)
		}
	default: // hybrid
		if ext != "" {
			byExt()
			break
		}
		cls.Sniffed = true
		mime, err := sniff(absPath)
		if err != nil {
			return cls, err
		}
		if mime == "" {
			cls.Group = GroupNoExtension
			cls.Text = false
		} else {
			cls.Group = mime
			cls.Text = c.MimeIsText(mime)
		}
	}
	return cls, nil
}

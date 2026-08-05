// Package main is the dirstat CLI entry point. It registers the scan command
// via strictcli and dispatches to the scan handler.
package main

import (
	"github.com/smm-h/strictcli/go/strictcli"
)

// newApp builds the fully-registered dirstat app. It is separate from main so
// tests can construct the same app and assert over its registration (see
// classification_test.go).
func newApp() *strictcli.App {
	app := strictcli.NewApp("dirstat", version,
		"Summarize files in a directory tree, grouped by format, with aggregate statistics")
	registerScanCmd(app)
	return app
}

func main() {
	newApp().Run()
}

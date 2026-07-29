// Package main is the dirstat CLI entry point. It registers the scan command
// via strictcli and dispatches to the scan handler.
package main

import (
	"github.com/smm-h/strictcli/go/strictcli"
)

func main() {
	app := strictcli.NewApp("dirstat", version,
		"Summarize files in a directory tree, grouped by format, with aggregate statistics")
	registerScanCmd(app)
	app.Run()
}

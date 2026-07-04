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

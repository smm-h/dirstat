package main

import (
	"runtime/debug"
	"strings"
)

// Version is set by ldflags at build time: -X main.Version=x.y.z
// The name must stay exported and spelled this way: .goreleaser.yml injects
// main.Version, and the linker silently does nothing when the symbol is absent.
var Version = ""

func init() {
	if Version != "" {
		Version = strings.TrimPrefix(Version, "v")
		return
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "(devel)" {
		Version = strings.TrimPrefix(info.Main.Version, "v")
	} else {
		Version = "dev"
	}
}

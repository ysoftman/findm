package main

import "runtime/debug"

// Version is set via ldflags at build time:
//
//	go build -ldflags "-X main.Version=1.0.0"
//
// Without ldflags it falls back to the module version recorded by
// `go install github.com/ysoftman/findm@v1.0.0`, otherwise "dev".
var Version = "dev"

func init() {
	if Version != "dev" {
		return
	}
	if bi, ok := debug.ReadBuildInfo(); ok && bi.Main.Version != "" && bi.Main.Version != "(devel)" {
		Version = bi.Main.Version
	}
}

package main

// Version is set via ldflags at build time:
//
//	go build -ldflags "-X main.Version=1.0.0"
//
// If not set, defaults to "dev".
var Version = "dev"

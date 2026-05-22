package main

import (
	"os"

	"github.com/vietnamesekid/usher/cmd"
)

// Set by goreleaser via -ldflags.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if err := cmd.NewRootCmd(version, commit, date).Execute(); err != nil {
		os.Exit(1)
	}
}

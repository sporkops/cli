package main

import (
	"os"

	"github.com/sporkops/cli/cmd"
	"github.com/sporkops/spork-go"
)

var version = "dev"

func main() {
	cmd.SetVersion(version)
	spork.Version = version
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

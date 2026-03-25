package main

import (
	"os"

	"github.com/sporkops/cli/cmd"
	"github.com/sporkops/cli/internal/api"
)

var version = "dev"

func main() {
	cmd.SetVersion(version)
	api.Version = version
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

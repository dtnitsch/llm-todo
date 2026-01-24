package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// version is set via -ldflags at build time
// Example: go build -ldflags "-X main.version=v0.6.0"
var version = "v0.7.0"

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Long: `Show the current version of llmtodo.

If built from source without version info, shows "dev".
Official releases include semantic version (e.g., v0.1.0).`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("llmtodo version %s\n", version)
		},
	}
}

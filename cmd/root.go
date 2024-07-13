package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "obsidian-cli",
	Short:   "obsidian-cli - CLI to open, search, move, create and update notes",
	Version: "v0.1.7",
	Long:    "obsidian-cli - CLI to open, search, move, create and update notes",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Whoops. There was an error while executing your CLI '%s'", err)
		os.Exit(1)
	}
}

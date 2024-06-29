package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "obsd",
	Short:   "obsd - CLI to open, search, move, create and update notes",
	Version: "v0.1.6",
	Long:    "obsd - CLI to open, search, move, create and update notes",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Whoops. There was an error while executing your CLI '%s'", err)
		os.Exit(1)
	}
}

package cmd

import "github.com/spf13/cobra"

var noteCmd = &cobra.Command{
	Use:   "note",
	Short: "Work with notes (create/open/print/rename/delete)",
}

func init() {
	rootCmd.AddCommand(noteCmd)
}

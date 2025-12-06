package cmd

import "github.com/spf13/cobra"

var fileCmd = &cobra.Command{
	Use:   "file",
	Short: "Alias for note commands (move/rename) that also work for attachments",
	RunE: func(cmd *cobra.Command, _ []string) error {
		return cmd.Help()
	},
}

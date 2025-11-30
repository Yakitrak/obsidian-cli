package cmd

import "github.com/spf13/cobra"

var vaultCmd = &cobra.Command{
	Use:   "vault",
	Short: "Manage vault selection and paths",
}

func init() {
	rootCmd.AddCommand(vaultCmd)
}

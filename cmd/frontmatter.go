package cmd

import "github.com/spf13/cobra"

var frontmatterCmd = &cobra.Command{
	Use:   "frontmatter",
	Short: "Manipulate note YAML frontmatter",
}

func init() {
	rootCmd.AddCommand(frontmatterCmd)
}

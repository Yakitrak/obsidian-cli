package cmd

import (
	"fmt"
	"log"

	"github.com/atomicobject/obsidian-cli/pkg/actions"
	"github.com/atomicobject/obsidian-cli/pkg/obsidian"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info [file]",
	Short: "Show file information including frontmatter and tags",
	Long: `Show detailed information about a file including its frontmatter and all tags.
Tags can be defined either in frontmatter or as hashtags in the file content.

Example:
  obsidian-cli info "Notes/Project.md"`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// If no vault name is provided, get the default vault name
		if vaultName == "" {
			vault := &obsidian.Vault{}
			defaultName, err := vault.DefaultName()
			if err != nil {
				log.Fatal(err)
			}
			vaultName = defaultName
		}

		vault := obsidian.Vault{Name: vaultName}
		note := obsidian.Note{}

		info, err := actions.GetFileInfo(&vault, &note, args[0])
		if err != nil {
			log.Fatal(err)
		}

		// Print the file information
		fmt.Println("File:", args[0])
		fmt.Println("\nFrontmatter:")
		if info.Frontmatter != nil {
			for k, v := range info.Frontmatter {
				fmt.Printf("  %s: %v\n", k, v)
			}
		} else {
			fmt.Println("  No frontmatter found")
		}

		fmt.Println("\nTags:")
		if len(info.Tags) > 0 {
			for _, tag := range info.Tags {
				fmt.Printf("  %s\n", tag)
			}
		} else {
			fmt.Println("  No tags found")
		}
	},
}

func init() {
	infoCmd.Flags().StringVarP(&vaultName, "vault", "v", "", "vault name")
	rootCmd.AddCommand(infoCmd)
}

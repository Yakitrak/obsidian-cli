package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Yakitrak/obsidian-cli/pkg/actions"
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"github.com/spf13/cobra"
)

var tagsCmd = &cobra.Command{
	Use:   "tags",
	Short: "List all tags in the vault",
	Long: `List all tags found in the vault, showing both individual counts (notes that contain this exact tag)
and aggregate counts (notes that contain this tag or any descendant tag).

Tags are sorted by aggregate count in descending order, with hierarchical tags grouped together.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonOutput, _ := cmd.Flags().GetBool("json")
		markdownOutput, _ := cmd.Flags().GetBool("markdown")

		// If no vault name is provided, get the default vault name
		if vaultName == "" {
			vault := &obsidian.Vault{}
			defaultName, err := vault.DefaultName()
			if err != nil {
				return fmt.Errorf("failed to get default vault name: %w", err)
			}
			vaultName = defaultName
		}

		// Get vault and note managers
		vault := obsidian.Vault{Name: vaultName}
		note := obsidian.Note{}

		// Get tags
		tagSummaries, err := actions.Tags(&vault, &note)
		if err != nil {
			return fmt.Errorf("failed to get tags: %w", err)
		}

		if jsonOutput {
			return outputTagsJSON(tagSummaries)
		}

		if markdownOutput {
			return outputTagsMarkdown(tagSummaries)
		}

		return outputTagsTable(tagSummaries)
	},
}

func outputTagsJSON(tagSummaries []actions.TagSummary) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(tagSummaries)
}

func outputTagsTable(tagSummaries []actions.TagSummary) error {
	if len(tagSummaries) == 0 {
		fmt.Println("No tags found in vault.")
		return nil
	}

	// Print header
	fmt.Printf("%-30s %6s %6s\n", "Tag", "Indiv", "Total")
	fmt.Printf("%-30s %6s %6s\n", "---", "-----", "-----")

	// Print each tag
	for _, tag := range tagSummaries {
		fmt.Printf("%-30s %6d %6d\n", tag.Name, tag.IndividualCount, tag.AggregateCount)
	}

	return nil
}

func outputTagsMarkdown(tagSummaries []actions.TagSummary) error {
	if len(tagSummaries) == 0 {
		fmt.Println("No tags found in vault.")
		return nil
	}

	// Print markdown table header
	fmt.Println("| Tag | Indiv | Total |")
	fmt.Println("|-----|-------|-------|")

	// Print each tag
	for _, tag := range tagSummaries {
		fmt.Printf("| %s | %d | %d |\n", tag.Name, tag.IndividualCount, tag.AggregateCount)
	}

	return nil
}

func init() {
	tagsCmd.Flags().StringVarP(&vaultName, "vault", "v", "", "vault name")
	tagsCmd.Flags().Bool("json", false, "Output tags as JSON")
	tagsCmd.Flags().Bool("markdown", false, "Output tags as markdown table")
	rootCmd.AddCommand(tagsCmd)
}

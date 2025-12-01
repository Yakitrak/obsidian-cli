package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/atomicobject/obsidian-cli/pkg/actions"
	"github.com/atomicobject/obsidian-cli/pkg/obsidian"
	"github.com/spf13/cobra"
)

var (
	tagsJSON         bool
	tagsMarkdown     bool
	tagsMatch        []string
	tagsMutationJSON bool
	tagsMutationMD   bool
	tagsDryRun       bool
	tagsWorkers      int
	tagsInputs       []string
	tagsRenameTarget string
)

var tagsCmd = &cobra.Command{
	Use:   "tags",
	Short: "Manage tags (list/add/delete/rename)",
	Long: `Manage tags in the vault using subcommands.

Examples:
  obscli tags list                           # List all tags
  obscli tags list --match tag:project       # List tags for project notes
  obscli tags add work urgent --inputs tag:project find:meeting
  obscli tags delete work urgent --dry-run
  obscli tags rename old --to new --workers 4`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		return cmd.Help()
	},
}

var listTagsCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List tags with individual and aggregate counts",
	RunE:    runListTags,
}

var addTagsCmd = &cobra.Command{
	Use:   "add <tag> [<tag>...] --inputs <criteria...>",
	Short: "Add tags to notes matching input criteria",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(tagsInputs) == 0 {
			return fmt.Errorf("--inputs is required (e.g., tag:project, find:meeting, paths)")
		}

		if err := ensureVaultName(); err != nil {
			return err
		}

		vault := obsidian.Vault{Name: vaultName}
		note := obsidian.Note{}

		inputs, expr, err := actions.ParseInputsWithExpression(tagsInputs)
		if err != nil {
			return fmt.Errorf("error parsing input criteria: %w", err)
		}

		matchingFiles, err := actions.ListFiles(&vault, &note, actions.ListParams{
			Inputs:         inputs,
			Expression:     expr,
			MaxDepth:       0,
			SkipAnchors:    false,
			SkipEmbeds:     false,
			AbsolutePaths:  false,
			SuppressedTags: []string{},
		})
		if err != nil {
			return fmt.Errorf("failed to get matching files: %w", err)
		}

		if len(matchingFiles) == 0 {
			fmt.Println("No files match the specified criteria.")
			return nil
		}

		summary, err := actions.AddTagsToFilesWithWorkers(&vault, &note, args, matchingFiles, tagsDryRun, tagsWorkers)
		if err != nil {
			return fmt.Errorf("failed to add tags: %w", err)
		}

		return outputMutationSummary(summary, "add", tagsDryRun, tagsMutationJSON, tagsMutationMD)
	},
}

var deleteTagsCmd = &cobra.Command{
	Use:     "delete <tag> [<tag>...]",
	Aliases: []string{"del", "rm"},
	Short:   "Delete tags from all notes that contain them",
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := ensureVaultName(); err != nil {
			return err
		}

		vault := obsidian.Vault{Name: vaultName}
		note := obsidian.Note{}

		summary, err := actions.DeleteTagsWithWorkers(&vault, &note, args, tagsDryRun, tagsWorkers)
		if err != nil {
			return fmt.Errorf("failed to delete tags: %w", err)
		}

		return outputMutationSummary(summary, "delete", tagsDryRun, tagsMutationJSON, tagsMutationMD)
	},
}

var renameTagsCmd = &cobra.Command{
	Use:   "rename <from-tag> [<from-tag>...] --to <to-tag>",
	Short: "Rename tags across the vault",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if strings.TrimSpace(tagsRenameTarget) == "" {
			return fmt.Errorf("--to destination tag is required")
		}

		if err := ensureVaultName(); err != nil {
			return err
		}

		vault := obsidian.Vault{Name: vaultName}
		note := obsidian.Note{}

		summary, err := actions.RenameTagsWithWorkers(&vault, &note, args, tagsRenameTarget, tagsDryRun, tagsWorkers)
		if err != nil {
			return fmt.Errorf("failed to rename tags: %w", err)
		}

		return outputMutationSummary(summary, "rename", tagsDryRun, tagsMutationJSON, tagsMutationMD)
	},
}

func runListTags(cmd *cobra.Command, _ []string) error {
	if err := ensureVaultName(); err != nil {
		return err
	}

	vault := obsidian.Vault{Name: vaultName}
	note := obsidian.Note{}

	scanNotes, err := resolveMatches(&vault, &note, tagsMatch)
	if err != nil {
		return err
	}
	if len(tagsMatch) > 0 && len(scanNotes) == 0 {
		fmt.Println("No files match the specified criteria.")
		return nil
	}

	tagSummaries, err := actions.Tags(&vault, &note, actions.TagsOptions{Notes: scanNotes})
	if err != nil {
		return fmt.Errorf("failed to get tags: %w", err)
	}

	if tagsJSON {
		return outputTagsJSON(tagSummaries)
	}

	if tagsMarkdown {
		return outputTagsMarkdown(tagSummaries)
	}

	return outputTagsTable(tagSummaries)
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

func outputMutationSummary(summary actions.TagMutationSummary, operation string, dryRun bool, jsonOutput bool, markdownOutput bool) error {
	if jsonOutput {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(summary)
	}

	if markdownOutput {
		return outputMutationSummaryMarkdown(summary, operation, dryRun)
	}

	return outputMutationSummaryTable(summary, operation, dryRun)
}

func outputMutationSummaryTable(summary actions.TagMutationSummary, operation string, dryRun bool) error {
	verb := operation + "d"
	if dryRun {
		verb = "would " + operation
	}

	if summary.NotesTouched == 0 {
		fmt.Printf("No tags %s.\n", verb)
		return nil
	}

	fmt.Printf("%s tags in %d note(s):\n", strings.ToUpper(string(verb[0]))+verb[1:], summary.NotesTouched)

	if len(summary.TagChanges) > 0 {
		fmt.Println("\nTag changes:")
		for tag, count := range summary.TagChanges {
			fmt.Printf("  %s: %d note(s)\n", tag, count)
		}
	}

	if !dryRun && len(summary.FilesChanged) > 0 {
		fmt.Printf("\nFiles modified: %d\n", len(summary.FilesChanged))
	}

	return nil
}

func outputMutationSummaryMarkdown(summary actions.TagMutationSummary, operation string, dryRun bool) error {
	verb := operation + "d"
	if dryRun {
		verb = "would " + operation
	}

	if summary.NotesTouched == 0 {
		fmt.Printf("No tags %s.\n", verb)
		return nil
	}

	fmt.Printf("## %s tags in %d note(s)\n\n", strings.ToUpper(string(verb[0]))+verb[1:], summary.NotesTouched)

	if len(summary.TagChanges) > 0 {
		fmt.Println("| Tag | Notes Changed |")
		fmt.Println("|-----|---------------|")
		for tag, count := range summary.TagChanges {
			fmt.Printf("| %s | %d |\n", tag, count)
		}
		fmt.Println()
	}

	if !dryRun && len(summary.FilesChanged) > 0 {
		fmt.Printf("**Files modified:** %d\n", len(summary.FilesChanged))
	}

	return nil
}

func ensureVaultName() error {
	if vaultName != "" {
		return nil
	}

	vault := &obsidian.Vault{}
	defaultName, err := vault.DefaultName()
	if err != nil {
		return fmt.Errorf("failed to get default vault name: %w", err)
	}
	vaultName = defaultName
	return nil
}

func addListFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&tagsJSON, "json", false, "Output tags as JSON")
	cmd.Flags().BoolVar(&tagsMarkdown, "markdown", false, "Output tags as markdown table")
	cmd.Flags().StringSliceVarP(&tagsMatch, "match", "m", nil, "Restrict listing to files matched by find/tag/path patterns (only for listing)")
}

func addMutationFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&tagsMutationJSON, "json", false, "Output results as JSON")
	cmd.Flags().BoolVar(&tagsMutationMD, "markdown", false, "Output results as markdown table")
	cmd.Flags().BoolVar(&tagsDryRun, "dry-run", false, "Show what would be changed without making changes")
	cmd.Flags().IntVarP(&tagsWorkers, "workers", "w", runtime.NumCPU(), "Number of parallel workers")
}

func init() {
	tagsCmd.PersistentFlags().StringVarP(&vaultName, "vault", "v", "", "vault name")

	addListFlags(listTagsCmd)

	addMutationFlags(addTagsCmd)
	addMutationFlags(deleteTagsCmd)
	addMutationFlags(renameTagsCmd)

	addTagsCmd.Flags().StringSliceVarP(&tagsInputs, "inputs", "i", nil, "Input criteria (find:/tag:/paths or boolean expressions) to select target notes")
	renameTagsCmd.Flags().StringVarP(&tagsRenameTarget, "to", "t", "", "Destination tag name for rename operation")

	tagsCmd.AddCommand(listTagsCmd)
	tagsCmd.AddCommand(addTagsCmd)
	tagsCmd.AddCommand(deleteTagsCmd)
	tagsCmd.AddCommand(renameTagsCmd)
	rootCmd.AddCommand(tagsCmd)
}

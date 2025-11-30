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

var tagsCmd = &cobra.Command{
	Use:   "tags",
	Short: "List, add, delete, or rename tags in the vault",
	Long: `Manage tags in the vault. By default, lists all tags found in the vault, showing both individual counts 
(notes that contain this exact tag) and aggregate counts (notes that contain this tag or any descendant tag).

Tags are sorted by aggregate count in descending order, with hierarchical tags grouped together.

You can also add, delete, or rename tags:
- Use --add to add tags to specific notes (requires input criteria)
- Use --delete to remove tags from all notes that have them
- Use --rename with --to to rename tags across all notes that have them
- Use --dry-run to preview changes without making them
- Use --workers to control parallelism (default: CPU count)

Examples:
  obscli tags                                    # List all tags
  obscli tags --add work,urgent tag:project     # Add 'work' and 'urgent' tags to notes tagged 'project'
  obscli tags --add important find:meeting      # Add 'important' tag to notes with 'meeting' in filename
  obscli tags --add done "Notes/Project.md"     # Add 'done' tag to specific file
  obscli tags --delete work urgent              # Delete 'work' and 'urgent' tags from all notes
  obscli tags --rename old --to new             # Rename 'old' tag to 'new'
  obscli tags --add urgent tag:project --dry-run # Preview adding 'urgent' tag to notes tagged 'project'
  obscli tags --delete work --workers 4         # Delete with 4 parallel workers`,
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonOutput, _ := cmd.Flags().GetBool("json")
		markdownOutput, _ := cmd.Flags().GetBool("markdown")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		workers, _ := cmd.Flags().GetInt("workers")
		matchPatterns, _ := cmd.Flags().GetStringSlice("match")

		deleteTags, _ := cmd.Flags().GetStringSlice("delete")
		renameTags, _ := cmd.Flags().GetStringSlice("rename")
		addTags, _ := cmd.Flags().GetStringSlice("add")
		toTag, _ := cmd.Flags().GetString("to")

		// Support space-separated tags as positional arguments for delete/rename
		// For --add, positional args are input criteria (tag:, find:, paths), not additional tags
		if len(args) > 0 {
			if len(deleteTags) > 0 {
				deleteTags = append(deleteTags, args...)
			} else if len(renameTags) > 0 {
				renameTags = append(renameTags, args...)
			}
			// NOTE: For --add, args are treated as input criteria, not additional tags
		}

		// Validate flag combinations
		operationCount := 0
		if len(deleteTags) > 0 {
			operationCount++
		}
		if len(renameTags) > 0 {
			operationCount++
		}
		if len(addTags) > 0 {
			operationCount++
		}

		if operationCount > 1 {
			return fmt.Errorf("cannot use --delete, --rename, and --add together")
		}

		if operationCount > 0 && len(matchPatterns) > 0 {
			return fmt.Errorf("--match is only supported when listing tags")
		}

		if len(renameTags) > 0 && toTag == "" {
			return fmt.Errorf("--rename requires --to destination tag")
		}

		if len(deleteTags) == 0 && len(renameTags) == 0 && toTag != "" {
			return fmt.Errorf("--to can only be used with --rename")
		}

		if len(addTags) > 0 && len(args) == 0 {
			return fmt.Errorf("--add requires input criteria (e.g., tag:project, find:meeting, file paths)")
		}

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

		// Handle delete operation
		if len(deleteTags) > 0 {
			summary, err := actions.DeleteTagsWithWorkers(&vault, &note, deleteTags, dryRun, workers)
			if err != nil {
				return fmt.Errorf("failed to delete tags: %w", err)
			}
			return outputMutationSummary(summary, "delete", dryRun, jsonOutput, markdownOutput)
		}

		// Handle rename operation
		if len(renameTags) > 0 {
			summary, err := actions.RenameTagsWithWorkers(&vault, &note, renameTags, toTag, dryRun, workers)
			if err != nil {
				return fmt.Errorf("failed to rename tags: %w", err)
			}
			return outputMutationSummary(summary, "rename", dryRun, jsonOutput, markdownOutput)
		}

		// Handle add operation
		if len(addTags) > 0 {
			// Parse input criteria to get matching files
			inputs, expr, err := actions.ParseInputsWithExpression(args)
			if err != nil {
				return fmt.Errorf("error parsing input criteria: %w", err)
			}

			// Get list of files matching the input criteria
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

			// Add tags to the specific matching files
			summary, err := actions.AddTagsToFilesWithWorkers(&vault, &note, addTags, matchingFiles, dryRun, workers)
			if err != nil {
				return fmt.Errorf("failed to add tags: %w", err)
			}
			return outputMutationSummary(summary, "add", dryRun, jsonOutput, markdownOutput)
		}

		// Default: list tags
		scanNotes, err := resolveMatches(&vault, &note, matchPatterns)
		if err != nil {
			return err
		}
		if len(matchPatterns) > 0 && len(scanNotes) == 0 {
			fmt.Println("No files match the specified criteria.")
			return nil
		}

		tagSummaries, err := actions.Tags(&vault, &note, actions.TagsOptions{Notes: scanNotes})
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

func init() {
	tagsCmd.Flags().StringVarP(&vaultName, "vault", "v", "", "vault name")
	tagsCmd.Flags().Bool("json", false, "Output tags as JSON")
	tagsCmd.Flags().Bool("markdown", false, "Output tags as markdown table")
	tagsCmd.Flags().StringSliceP("delete", "d", nil, "Delete specified tags from all notes")
	tagsCmd.Flags().StringSliceP("rename", "r", nil, "Rename specified tags (use with --to)")
	tagsCmd.Flags().StringSliceP("add", "a", nil, "Add specified tags to matching notes (requires input criteria)")
	tagsCmd.Flags().StringP("to", "t", "", "Destination tag name for rename operation")
	tagsCmd.Flags().Bool("dry-run", false, "Show what would be changed without making changes")
	tagsCmd.Flags().IntP("workers", "w", runtime.NumCPU(), "Number of parallel workers")
	tagsCmd.Flags().StringSliceP("match", "m", nil, "Restrict listing to files matched by find/tag/path patterns (only for listing)")
	rootCmd.AddCommand(tagsCmd)
}

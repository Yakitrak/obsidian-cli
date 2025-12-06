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
	"gopkg.in/yaml.v3"
)

var propertiesCmd = &cobra.Command{
	Use:   "properties",
	Short: "Inspect or edit frontmatter properties across the vault",
	Long: `Scan frontmatter to show how properties are used across the vault, including value shapes, types, and enum-like values.

Tags are included by default; use --exclude-tags to skip them.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func outputPropertiesJSON(summaries []actions.PropertySummary) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(summaries)
}

func outputPropertiesTable(summaries []actions.PropertySummary) error {
	if len(summaries) == 0 {
		fmt.Println("No properties found in vault.")
		return nil
	}

	fmt.Printf("%-24s %6s %-8s %-10s %s\n", "Property", "Notes", "Shape", "Type", "Details")
	fmt.Printf("%-24s %6s %-8s %-10s %s\n", strings.Repeat("-", 24), strings.Repeat("-", 6), strings.Repeat("-", 8), strings.Repeat("-", 10), strings.Repeat("-", 30))

	for _, p := range summaries {
		details := propertyDetails(p)
		fmt.Printf("%-24s %6d %-8s %-10s %s\n", p.Name, p.NoteCount, p.Shape, p.ValueType, details)
	}

	return nil
}

func outputPropertiesMarkdown(summaries []actions.PropertySummary) error {
	if len(summaries) == 0 {
		fmt.Println("No properties found in vault.")
		return nil
	}

	fmt.Println("| Property | Notes | Shape | Type | Details |")
	fmt.Println("|----------|-------|-------|------|---------|")
	for _, p := range summaries {
		details := propertyDetails(p)
		fmt.Printf("| %s | %d | %s | %s | %s |\n", p.Name, p.NoteCount, p.Shape, p.ValueType, details)
	}
	return nil
}

func propertyDetails(p actions.PropertySummary) string {
	if len(p.EnumValues) > 0 {
		if len(p.EnumValueCounts) > 0 {
			items := make([]string, 0, len(p.EnumValues))
			for _, v := range p.EnumValues {
				items = append(items, fmt.Sprintf("%s(%d)", v, p.EnumValueCounts[v]))
			}
			return strings.Join(items, ", ")
		}
		return strings.Join(p.EnumValues, ", ")
	}

	if p.DistinctValueCount == 0 {
		return "no values"
	}

	if p.TruncatedValueSet {
		return fmt.Sprintf("%d+ distinct", p.DistinctValueCount)
	}

	return fmt.Sprintf("%d distinct", p.DistinctValueCount)
}

func init() {
	propertiesCmd.PersistentFlags().StringVarP(&vaultName, "vault", "v", "", "vault name")
	listPropertiesCmd.Flags().Bool("json", false, "Output properties as JSON")
	listPropertiesCmd.Flags().Bool("markdown", false, "Output properties as markdown table")
	listPropertiesCmd.Flags().Bool("exclude-tags", false, "Exclude the tags frontmatter field")
	listPropertiesCmd.Flags().Int("value-limit", 5, "Emit values when distinct counts are at or below this limit")
	listPropertiesCmd.Flags().Int("enum-threshold", 5, "Deprecated: use --value-limit")
	listPropertiesCmd.Flags().Int("max-values", 500, "Maximum distinct values to track per property for reporting")
	listPropertiesCmd.Flags().StringSlice("only", nil, "Only include these property names (repeatable)")
	listPropertiesCmd.Flags().StringSliceP("match", "m", nil, "Restrict analysis to files matched by find/tag/path patterns")
	listPropertiesCmd.Flags().Bool("verbose", false, "Show expanded enums: allow mixed-type enums and raise value limit to 50")
	listPropertiesCmd.Flags().Bool("value-counts", false, "Include per-value note counts for value outputs")
	listPropertiesCmd.Flags().Bool("enum-counts", false, "Deprecated: use --value-counts")
	listPropertiesCmd.Flags().String("source", "all", "Property source to scan: all, frontmatter, or inline")
	listPropertiesCmd.Flags().MarkDeprecated("enum-threshold", "use --value-limit")
	listPropertiesCmd.Flags().MarkDeprecated("enum-counts", "use --value-counts")

	propertiesSetCmd.Flags().StringSliceVar(&propertiesInputs, "inputs", nil, "Input criteria (find:, tag:, or paths) to scope property setting")
	propertiesSetCmd.Flags().StringVar(&propertiesValue, "value", "", "Property value to set (YAML accepted)")
	propertiesSetCmd.Flags().BoolVar(&propertiesOverwrite, "overwrite", false, "Overwrite existing values")
	propertiesSetCmd.Flags().BoolVar(&propertiesDryRun, "dry-run", false, "Preview changes without writing files")
	propertiesSetCmd.Flags().IntVar(&propertiesWorkers, "workers", runtime.NumCPU(), "Number of workers to use")
	propertiesSetCmd.Flags().BoolVar(&propertiesMutationJSON, "json", false, "Output mutation summary as JSON")
	propertiesSetCmd.Flags().BoolVar(&propertiesMutationMD, "markdown", false, "Output mutation summary as markdown")

	propertiesDeleteCmd.Flags().StringSliceVar(&propertiesInputs, "inputs", nil, "Optional input criteria (find:, tag:, or paths) to scope deletion")
	propertiesDeleteCmd.Flags().BoolVar(&propertiesDryRun, "dry-run", false, "Preview changes without writing files")
	propertiesDeleteCmd.Flags().IntVar(&propertiesWorkers, "workers", runtime.NumCPU(), "Number of workers to use")
	propertiesDeleteCmd.Flags().BoolVar(&propertiesMutationJSON, "json", false, "Output mutation summary as JSON")
	propertiesDeleteCmd.Flags().BoolVar(&propertiesMutationMD, "markdown", false, "Output mutation summary as markdown")

	propertiesRenameCmd.Flags().StringSliceVar(&propertiesInputs, "inputs", nil, "Optional input criteria (find:, tag:, or paths) to scope rename")
	propertiesRenameCmd.Flags().StringVar(&propertiesRenameTarget, "to", "", "Destination property name")
	propertiesRenameCmd.Flags().BoolVar(&propertiesMerge, "merge", true, "Merge values into destination when it already exists")
	propertiesRenameCmd.Flags().BoolVar(&propertiesDryRun, "dry-run", false, "Preview changes without writing files")
	propertiesRenameCmd.Flags().IntVar(&propertiesWorkers, "workers", runtime.NumCPU(), "Number of workers to use")
	propertiesRenameCmd.Flags().BoolVar(&propertiesMutationJSON, "json", false, "Output mutation summary as JSON")
	propertiesRenameCmd.Flags().BoolVar(&propertiesMutationMD, "markdown", false, "Output mutation summary as markdown")

	propertiesCmd.AddCommand(listPropertiesCmd)
	propertiesCmd.AddCommand(propertiesSetCmd)
	propertiesCmd.AddCommand(propertiesDeleteCmd)
	propertiesCmd.AddCommand(propertiesRenameCmd)
	rootCmd.AddCommand(propertiesCmd)
}

var (
	propertiesMutationJSON bool
	propertiesMutationMD   bool
	propertiesDryRun       bool
	propertiesWorkers      int
	propertiesInputs       []string
	propertiesValue        string
	propertiesOverwrite    bool
	propertiesRenameTarget string
	propertiesMerge        bool = true
)

var propertiesSetCmd = &cobra.Command{
	Use:   "set <property> --value <yaml> --inputs <criteria...>",
	Short: "Set a frontmatter property on matching notes",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(propertiesInputs) == 0 {
			return fmt.Errorf("--inputs is required (e.g., tag:project, find:meeting, paths)")
		}
		if strings.TrimSpace(propertiesValue) == "" {
			return fmt.Errorf("--value is required")
		}
		if err := ensureVaultName(); err != nil {
			return err
		}

		vault := obsidian.Vault{Name: vaultName}
		note := obsidian.Note{}

		inputs, expr, err := actions.ParseInputsWithExpression(propertiesInputs)
		if err != nil {
			return fmt.Errorf("error parsing input criteria: %w", err)
		}

		matchingFiles, err := actions.ListFiles(&vault, &note, actions.ListParams{
			Inputs:        inputs,
			Expression:    expr,
			MaxDepth:      0,
			SkipAnchors:   false,
			SkipEmbeds:    false,
			AbsolutePaths: false,
		})
		if err != nil {
			return fmt.Errorf("failed to get matching files: %w", err)
		}
		if len(matchingFiles) == 0 {
			fmt.Println("No files match the specified criteria.")
			return nil
		}

		var parsed interface{}
		if err := yaml.Unmarshal([]byte(propertiesValue), &parsed); err != nil {
			parsed = propertiesValue
		}

		summary, err := actions.SetPropertyOnFilesWithWorkers(&vault, &note, args[0], parsed, matchingFiles, propertiesOverwrite, propertiesDryRun, propertiesWorkers)
		if err != nil {
			return fmt.Errorf("failed to set property: %w", err)
		}

		return outputPropertyMutationSummary(summary, "set", propertiesDryRun, propertiesMutationJSON, propertiesMutationMD)
	},
}

var propertiesDeleteCmd = &cobra.Command{
	Use:   "delete <property> [<property>...]",
	Short: "Delete frontmatter properties across the vault (or scoped inputs)",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := ensureVaultName(); err != nil {
			return err
		}

		vault := obsidian.Vault{Name: vaultName}
		note := obsidian.Note{}

		var files []string
		if len(propertiesInputs) > 0 {
			inputs, expr, err := actions.ParseInputsWithExpression(propertiesInputs)
			if err != nil {
				return fmt.Errorf("error parsing input criteria: %w", err)
			}
			files, err = actions.ListFiles(&vault, &note, actions.ListParams{
				Inputs:        inputs,
				Expression:    expr,
				MaxDepth:      0,
				SkipAnchors:   false,
				SkipEmbeds:    false,
				AbsolutePaths: false,
			})
			if err != nil {
				return fmt.Errorf("failed to get matching files: %w", err)
			}
			if len(files) == 0 {
				fmt.Println("No files match the specified criteria.")
				return nil
			}
		}

		summary, err := actions.DeletePropertiesWithWorkers(&vault, &note, args, files, propertiesDryRun, propertiesWorkers)
		if err != nil {
			return fmt.Errorf("failed to delete properties: %w", err)
		}

		return outputPropertyMutationSummary(summary, "delete", propertiesDryRun, propertiesMutationJSON, propertiesMutationMD)
	},
}

var propertiesRenameCmd = &cobra.Command{
	Use:   "rename <from> [<from>...] --to <to>",
	Short: "Rename (and optionally merge) frontmatter properties",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if strings.TrimSpace(propertiesRenameTarget) == "" {
			return fmt.Errorf("--to destination property is required")
		}
		if err := ensureVaultName(); err != nil {
			return err
		}

		vault := obsidian.Vault{Name: vaultName}
		note := obsidian.Note{}

		var files []string
		if len(propertiesInputs) > 0 {
			inputs, expr, err := actions.ParseInputsWithExpression(propertiesInputs)
			if err != nil {
				return fmt.Errorf("error parsing input criteria: %w", err)
			}
			files, err = actions.ListFiles(&vault, &note, actions.ListParams{
				Inputs:        inputs,
				Expression:    expr,
				MaxDepth:      0,
				SkipAnchors:   false,
				SkipEmbeds:    false,
				AbsolutePaths: false,
			})
			if err != nil {
				return fmt.Errorf("failed to get matching files: %w", err)
			}
			if len(files) == 0 {
				fmt.Println("No files match the specified criteria.")
				return nil
			}
		}

		summary, err := actions.RenamePropertiesWithWorkers(&vault, &note, args, propertiesRenameTarget, propertiesMerge, files, propertiesDryRun, propertiesWorkers)
		if err != nil {
			return fmt.Errorf("failed to rename properties: %w", err)
		}

		return outputPropertyMutationSummary(summary, "rename", propertiesDryRun, propertiesMutationJSON, propertiesMutationMD)
	},
}

func outputPropertyMutationSummary(summary actions.PropertyMutationSummary, operation string, dryRun bool, jsonOutput bool, markdownOutput bool) error {
	if jsonOutput {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(summary)
	}

	if markdownOutput {
		return outputPropertyMutationSummaryMarkdown(summary, operation, dryRun)
	}

	return outputPropertyMutationSummaryTable(summary, operation, dryRun)
}

func outputPropertyMutationSummaryTable(summary actions.PropertyMutationSummary, operation string, dryRun bool) error {
	verb := operation + "d"
	if dryRun {
		verb = "would " + operation
	}

	if summary.NotesTouched == 0 {
		fmt.Printf("No properties %s.\n", verb)
		return nil
	}

	fmt.Printf("%s properties in %d note(s):\n", strings.ToUpper(string(verb[0]))+verb[1:], summary.NotesTouched)

	if len(summary.PropertyChanges) > 0 {
		fmt.Println("\nProperty changes:")
		for prop, count := range summary.PropertyChanges {
			fmt.Printf("  %s: %d note(s)\n", prop, count)
		}
	}

	if !dryRun && len(summary.FilesChanged) > 0 {
		fmt.Printf("\nFiles modified: %d\n", len(summary.FilesChanged))
	}

	return nil
}

func outputPropertyMutationSummaryMarkdown(summary actions.PropertyMutationSummary, operation string, dryRun bool) error {
	verb := operation + "d"
	if dryRun {
		verb = "would " + operation
	}

	if summary.NotesTouched == 0 {
		fmt.Printf("No properties %s.\n", verb)
		return nil
	}

	fmt.Printf("## %s properties in %d note(s)\n\n", strings.ToUpper(string(verb[0]))+verb[1:], summary.NotesTouched)

	if len(summary.PropertyChanges) > 0 {
		fmt.Println("| Property | Notes Changed |")
		fmt.Println("|----------|---------------|")
		for prop, count := range summary.PropertyChanges {
			fmt.Printf("| %s | %d |\n", prop, count)
		}
		fmt.Println()
	}

	if !dryRun && len(summary.FilesChanged) > 0 {
		fmt.Printf("**Files modified:** %d\n", len(summary.FilesChanged))
	}

	return nil
}

var listPropertiesCmd = &cobra.Command{
	Use:   "list",
	Short: "List frontmatter/inline properties across the vault",
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonOutput, _ := cmd.Flags().GetBool("json")
		markdownOutput, _ := cmd.Flags().GetBool("markdown")
		excludeTags, _ := cmd.Flags().GetBool("exclude-tags")
		valueLimit, _ := cmd.Flags().GetInt("value-limit")
		if !cmd.Flags().Changed("value-limit") && cmd.Flags().Changed("enum-threshold") {
			valueLimit, _ = cmd.Flags().GetInt("enum-threshold")
		}
		maxValues, _ := cmd.Flags().GetInt("max-values")
		onlyProps, _ := cmd.Flags().GetStringSlice("only")
		matchPatterns, _ := cmd.Flags().GetStringSlice("match")
		verboseEnums, _ := cmd.Flags().GetBool("verbose")
		includeValueCounts, _ := cmd.Flags().GetBool("value-counts")
		if !cmd.Flags().Changed("value-counts") && cmd.Flags().Changed("enum-counts") {
			includeValueCounts, _ = cmd.Flags().GetBool("enum-counts")
		}
		sourceFlag, _ := cmd.Flags().GetString("source")

		var source actions.PropertySource
		switch sourceFlag {
		case "", "all":
			source = actions.PropertySourceAll
		case "frontmatter":
			source = actions.PropertySourceFrontmatter
		case "inline":
			source = actions.PropertySourceInline
		default:
			return fmt.Errorf("invalid --source value %q: must be all, frontmatter, or inline", sourceFlag)
		}

		if vaultName == "" {
			vault := &obsidian.Vault{}
			defaultName, err := vault.DefaultName()
			if err != nil {
				return fmt.Errorf("failed to get default vault name: %w", err)
			}
			vaultName = defaultName
		}

		vault := obsidian.Vault{Name: vaultName}
		note := obsidian.Note{}

		scanNotes, err := resolveMatches(&vault, &note, matchPatterns)
		if err != nil {
			return err
		}
		if len(matchPatterns) > 0 && len(scanNotes) == 0 {
			fmt.Println("No files match the specified criteria.")
			return nil
		}

		valueLimitChanged := cmd.Flags().Changed("value-limit") || cmd.Flags().Changed("enum-threshold")
		if len(onlyProps) > 0 && !valueLimitChanged {
			if maxValues <= 0 {
				maxValues = 500
			}
			if maxValues > 1 {
				valueLimit = maxValues - 1
			} else {
				valueLimit = maxValues
			}
		}

		if verboseEnums && valueLimit < 50 {
			valueLimit = 50
		}
		if maxValues < valueLimit+1 {
			maxValues = valueLimit + 1
		}

		summaries, err := actions.Properties(&vault, &note, actions.PropertiesOptions{
			ExcludeTags:        excludeTags,
			ValueLimit:         valueLimit,
			MaxValues:          maxValues,
			Notes:              scanNotes,
			Only:               onlyProps,
			ForceEnumMixed:     verboseEnums,
			Source:             source,
			IncludeValueCounts: includeValueCounts,
		})
		if err != nil {
			return fmt.Errorf("failed to analyze properties: %w", err)
		}

		if jsonOutput {
			return outputPropertiesJSON(summaries)
		}

		if markdownOutput {
			return outputPropertiesMarkdown(summaries)
		}

		return outputPropertiesTable(summaries)
	},
}

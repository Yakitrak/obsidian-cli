package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/atomicobject/obsidian-cli/pkg/actions"
	"github.com/atomicobject/obsidian-cli/pkg/obsidian"
	"github.com/spf13/cobra"
)

var propertiesCmd = &cobra.Command{
	Use:   "properties",
	Short: "Inspect frontmatter properties across the vault",
	Long: `Scan frontmatter to show how properties are used across the vault, including value shapes, types, and enum-like values.

Tags are included by default; use --exclude-tags to skip them.`,
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

		// Validate source flag
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
	propertiesCmd.Flags().StringVarP(&vaultName, "vault", "v", "", "vault name")
	propertiesCmd.Flags().Bool("json", false, "Output properties as JSON")
	propertiesCmd.Flags().Bool("markdown", false, "Output properties as markdown table")
	propertiesCmd.Flags().Bool("exclude-tags", false, "Exclude the tags frontmatter field")
	propertiesCmd.Flags().Int("value-limit", 5, "Emit values when distinct counts are at or below this limit")
	propertiesCmd.Flags().Int("enum-threshold", 5, "Deprecated: use --value-limit")
	propertiesCmd.Flags().Int("max-values", 500, "Maximum distinct values to track per property for reporting")
	propertiesCmd.Flags().StringSlice("only", nil, "Only include these property names (repeatable)")
	propertiesCmd.Flags().StringSliceP("match", "m", nil, "Restrict analysis to files matched by find/tag/path patterns")
	propertiesCmd.Flags().Bool("verbose", false, "Show expanded enums: allow mixed-type enums and raise value limit to 50")
	propertiesCmd.Flags().Bool("value-counts", false, "Include per-value note counts for value outputs")
	propertiesCmd.Flags().Bool("enum-counts", false, "Deprecated: use --value-counts")
	propertiesCmd.Flags().String("source", "all", "Property source to scan: all, frontmatter, or inline")
	propertiesCmd.Flags().MarkDeprecated("enum-threshold", "use --value-limit")
	propertiesCmd.Flags().MarkDeprecated("enum-counts", "use --value-counts")
	rootCmd.AddCommand(propertiesCmd)
}

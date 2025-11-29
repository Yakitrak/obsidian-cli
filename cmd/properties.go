package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/Yakitrak/obsidian-cli/pkg/actions"
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
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
		enumThreshold, _ := cmd.Flags().GetInt("enum-threshold")
		maxValues, _ := cmd.Flags().GetInt("max-values")
		matchPatterns, _ := cmd.Flags().GetStringSlice("match")
		verboseEnums, _ := cmd.Flags().GetBool("verbose")
		includeValueCounts, _ := cmd.Flags().GetBool("enum-counts")

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

		if verboseEnums && enumThreshold < 50 {
			enumThreshold = 50
		}

		summaries, err := actions.Properties(&vault, &note, actions.PropertiesOptions{
			ExcludeTags:        excludeTags,
			EnumThreshold:      enumThreshold,
			MaxValues:          maxValues,
			Notes:              scanNotes,
			ForceEnumMixed:     verboseEnums,
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
	propertiesCmd.Flags().Int("enum-threshold", 5, "Emit enum values when distinct values are at or below this count")
	propertiesCmd.Flags().Int("max-values", 500, "Maximum distinct values to track per property for reporting")
	propertiesCmd.Flags().StringSliceP("match", "m", nil, "Restrict analysis to files matched by find/tag/path patterns")
	propertiesCmd.Flags().Bool("verbose", false, "Show expanded enums: allow mixed-type enums and raise enum threshold to 50")
	propertiesCmd.Flags().Bool("enum-counts", false, "Include per-value note counts for enum outputs")
	rootCmd.AddCommand(propertiesCmd)
}

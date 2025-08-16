package cmd

import (
	"log"
	"strings"

	"github.com/Yakitrak/obsidian-cli/pkg/actions"
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"github.com/spf13/cobra"
)

var fmClearKey string

var frontmatterClearCmd = &cobra.Command{
	Use:   "clear <note>",
	Short: "Clear the content of a YAML frontmatter key in a note",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		v := obsidian.Vault{Name: vaultName}
		note := args[0]
		for _, k := range parseKeyList(fmClearKey) {
			if err := actions.ClearFrontmatter(&v, note, k); err != nil {
				log.Fatal(err)
			}
		}
	},
}

// parseKeyList accepts a single key, a comma-separated list ("a, b, c"), or a bracketed list ("[a, b, c]")
// and returns a slice of trimmed, non-empty keys.
func parseKeyList(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	// Strip surrounding brackets if present
	if strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]") {
		s = strings.TrimSpace(s[1 : len(s)-1])
	}
	// If no comma, return single key
	if !strings.Contains(s, ",") {
		if s == "" {
			return nil
		}
		return []string{s}
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func init() {
	frontmatterClearCmd.Flags().StringVarP(&vaultName, "vault", "v", "", "vault name (not required if default is set)")
	frontmatterClearCmd.Flags().StringVarP(&fmClearKey, "key", "k", "", "frontmatter key(s) to clear (supports \"a, b\" or \"[a, b]\")")
	if err := frontmatterClearCmd.MarkFlagRequired("key"); err != nil {
		log.Fatalf("failed to mark --key as required: %v", err)
	}
	frontmatterCmd.AddCommand(frontmatterClearCmd)
}

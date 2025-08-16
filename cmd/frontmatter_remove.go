package cmd

import (
	"log"
	"strings"

	"github.com/Yakitrak/obsidian-cli/pkg/actions"
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"github.com/spf13/cobra"
)

var fmRemoveKey string

var frontmatterRemoveCmd = &cobra.Command{
	Use:   "remove <note>",
	Short: "Remove a YAML frontmatter key (and its value) from a note",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		v := obsidian.Vault{Name: vaultName}
		note := args[0]
		for _, k := range parseKeyListRemove(fmRemoveKey) {
			if err := actions.RemoveFrontmatterKey(&v, note, k); err != nil {
				log.Fatal(err)
			}
		}
	},
}

// parseKeyListRemove accepts a single key, a comma-separated list ("a, b, c"), or a bracketed list ("[a, b, c]")
// and returns a slice of trimmed, non-empty keys.
func parseKeyListRemove(s string) []string {
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
	frontmatterRemoveCmd.Flags().StringVarP(&vaultName, "vault", "v", "", "vault name (not required if default is set)")
	frontmatterRemoveCmd.Flags().StringVarP(&fmRemoveKey, "key", "k", "", "frontmatter key(s) to remove (supports \"a, b\" or \"[a, b]\")")
	if err := frontmatterRemoveCmd.MarkFlagRequired("key"); err != nil {
		log.Fatalf("failed to mark --key as required: %v", err)
	}
	frontmatterCmd.AddCommand(frontmatterRemoveCmd)
}

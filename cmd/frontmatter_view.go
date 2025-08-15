package cmd

import (
	"fmt"
	"log"

	"github.com/Yakitrak/obsidian-cli/pkg/actions"
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var fmViewKey string
var fmViewExpect string

var frontmatterViewCmd = &cobra.Command{
	Use:   "view <note>",
	Short: "View YAML frontmatter key in a note or test equality/containment",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		v := obsidian.Vault{Name: vaultName}
		note := args[0]

		fv, err := actions.GetFrontmatterValue(&v, note, fmViewKey)
		if err != nil {
			log.Fatal(err)
		}

		// If value flag provided, print true/false according to match rules
		if fmViewExpect != "" {
			fmt.Println(matchFrontmatterValue(fv, fmViewExpect))
			return
		}

		// Otherwise, print the actual value (YAML-serialized) if found
		if !fv.Found || fv.Value == nil {
			// print nothing on missing key
			return
		}
		// Pretty print the value in YAML but as a single scalar/sequence map dump (no frontmatter delimiters)
		out, err := yaml.Marshal(fv.Value)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Print(string(out))
	},
}

// matchFrontmatterValue evaluates if the actual value matches the expected string
// Rules:
// - If actual is a slice/array, returns true if any element stringifies exactly equals expected
// - If actual is a scalar, returns true if its stringified form equals expected
func matchFrontmatterValue(fv *actions.FrontmatterValue, expected string) bool {
	if fv == nil || !fv.Found || fv.Value == nil {
		return false
	}
	switch v := fv.Value.(type) {
	case []interface{}:
		for _, el := range v {
			if stringifyFM(el) == expected {
				return true
			}
		}
		return false
	case []string:
		for _, el := range v {
			if el == expected {
				return true
			}
		}
		return false
	default:
		return stringifyFM(v) == expected
	}
}

func stringifyFM(v interface{}) string {
	// Try YAML marshal then trim trailing newline
	b, err := yaml.Marshal(v)
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	// Remove any trailing newline introduced by yaml.Marshal
	s := string(b)
	if len(s) > 0 && s[len(s)-1] == '\n' {
		s = s[:len(s)-1]
	}
	return s
}

func init() {
	frontmatterViewCmd.Flags().StringVarP(&vaultName, "vault", "v", "", "vault name (not required if default is set)")
	frontmatterViewCmd.Flags().StringVarP(&fmViewKey, "key", "k", "", "frontmatter key to view")
	frontmatterViewCmd.Flags().StringVarP(&fmViewExpect, "value", "V", "", "expected value for boolean check")
	frontmatterViewCmd.MarkFlagRequired("key")
	frontmatterCmd.AddCommand(frontmatterViewCmd)
}

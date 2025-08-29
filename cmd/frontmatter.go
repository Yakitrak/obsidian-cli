package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Yakitrak/obsidian-cli/pkg/actions"
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	fmFlagEdit   bool
	fmFlagDelete bool
	fmFlagPrint  bool
	fmFlagClear  bool
	fmKeyFlag    string
	fmValueFlag  string
)

var frontmatterCmd = &cobra.Command{
	Use:   "frontmatter [--edit|--delete|--print|--clear] <note> [--key <k>] [--value <v>]",
	Short: "Manipulate note YAML frontmatter",
	Args:  cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Determine which action flag is set
		actionsSet := 0
		if fmFlagEdit {
			actionsSet++
		}
		if fmFlagDelete {
			actionsSet++
		}
		if fmFlagPrint {
			actionsSet++
		}
		if fmFlagClear {
			actionsSet++
		}
		if actionsSet == 0 {
			// No action flag: show help
			_ = cmd.Help()
			return errors.New("one of --edit, --delete, --print, or --clear must be specified")
		}
		if actionsSet > 1 {
			return errors.New("only one of --edit, --delete, --print, or --clear can be used at a time")
		}

		if len(args) != 1 {
			return errors.New("exactly one note argument is required")
		}
		note := args[0]
		v := obsidian.Vault{Name: vaultName}

		// Dispatch by action
		if fmFlagEdit {
			if fmKeyFlag == "" {
				return errors.New("--key is required for --edit")
			}
			// Support adding multiple empty keys at once using comma-separated --key when --value is omitted
			if fmValueFlag == "" && strings.Contains(fmKeyFlag, ",") {
				parts := strings.Split(fmKeyFlag, ",")
				for _, p := range parts {
					k := strings.TrimSpace(p)
					if k == "" {
						continue
					}
					params := actions.FrontmatterEditParams{NoteName: note, Key: k, Value: ""}
					if err := actions.EditFrontmatter(&v, params); err != nil {
						return err
					}
				}
				return nil
			}
			params := actions.FrontmatterEditParams{NoteName: note, Key: fmKeyFlag, Value: fmValueFlag}
			return actions.EditFrontmatter(&v, params)
		}

		if fmFlagDelete {
			if fmKeyFlag == "" {
				return errors.New("--key is required for --delete")
			}
			for _, k := range parseKeyListRemove(fmKeyFlag) {
				if err := actions.RemoveFrontmatterKey(&v, note, k); err != nil {
					return err
				}
			}
			return nil
		}

		if fmFlagClear {
			if fmKeyFlag == "" {
				return errors.New("--key is required for --clear")
			}
			for _, k := range parseKeyList(fmKeyFlag) {
				if err := actions.ClearFrontmatter(&v, note, k); err != nil {
					return err
				}
			}
			return nil
		}

		// fmFlagPrint
		if fmKeyFlag == "" {
			return errors.New("--key is required for --print")
		}
		fv, err := actions.GetFrontmatterValue(&v, note, fmKeyFlag)
		if err != nil {
			return err
		}
		if fmValueFlag != "" {
			fmt.Println(matchFrontmatterValue(fv, fmValueFlag))
			return nil
		}
		if !fv.Found || fv.Value == nil {
			return nil // print nothing on missing key
		}
		out, err := yaml.Marshal(fv.Value)
		if err != nil {
			return err
		}
		fmt.Print(string(out))
		return nil
	},
}

func init() {
	frontmatterCmd.Flags().StringVarP(&vaultName, "vault", "v", "", "vault name (not required if default is set)")
	frontmatterCmd.Flags().BoolVar(&fmFlagEdit, "edit", false, "edit YAML frontmatter key in a note")
	frontmatterCmd.Flags().BoolVar(&fmFlagDelete, "delete", false, "delete YAML frontmatter key(s) from a note")
	frontmatterCmd.Flags().BoolVar(&fmFlagPrint, "print", false, "print YAML frontmatter key value or boolean check")
	frontmatterCmd.Flags().BoolVar(&fmFlagClear, "clear", false, "clear YAML frontmatter key value(s) in a note")
	frontmatterCmd.Flags().StringVarP(&fmKeyFlag, "key", "k", "", "frontmatter key(s)")
	frontmatterCmd.Flags().StringVarP(&fmValueFlag, "value", "V", "", "frontmatter value (YAML) or expected value for --print")

	rootCmd.AddCommand(frontmatterCmd)
}

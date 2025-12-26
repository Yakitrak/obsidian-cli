package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/Yakitrak/obsidian-cli/pkg/actions"
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"github.com/spf13/cobra"
)

var (
	appendTimestamp bool
	appendTimeFmt   string
	appendDryRun    bool
)

var appendCmd = &cobra.Command{
	Use:     "append [text]",
	Aliases: []string{"a"},
	Short:   "Append text to today's daily note",
	Long: `Appends text to today's daily note.

This command writes to a daily note path derived from your per-vault settings
in preferences.json (daily_note.folder and daily_note.filename_pattern).

If no text argument is provided, content is read from stdin (piped) or entered
interactively until EOF.`,
	Example: `  # Append a one-liner
  obsidian-cli append "Meeting notes: discussed roadmap"

  # Append multi-line content interactively (Ctrl-D to save)
  obsidian-cli append

  # Append with timestamp
  obsidian-cli append --timestamp "Started work on feature X"

  # Append in a specific vault
  obsidian-cli append --vault "Work" "Daily standup notes"`,
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		vault := obsidian.Vault{Name: vaultName}

		content := strings.TrimSpace(strings.Join(args, " "))
		if content == "" {
			stat, err := os.Stdin.Stat()
			if err == nil && (stat.Mode()&os.ModeCharDevice) == 0 {
				b, err := io.ReadAll(os.Stdin)
				if err != nil {
					return err
				}
				content = strings.TrimSpace(string(b))
			} else if !appendDryRun {
				var err error
				content, err = actions.PromptForContentIfEmpty(content)
				if err != nil {
					return err
				}
			}
		}

		if appendTimestamp {
			format := appendTimeFmt
			if strings.TrimSpace(format) == "" {
				format = "15:04"
			}
			content = fmt.Sprintf("- %s %s", time.Now().Format(format), content)
		}

		if appendDryRun {
			plan, err := actions.PlanDailyAppend(&vault, time.Now())
			if err != nil {
				return err
			}
			fmt.Println("Append dry-run:")
			fmt.Printf("  vault: %s\n", plan.VaultName)
			fmt.Printf("  note: %s\n", plan.AbsoluteNotePath)
			fmt.Printf("  create_dirs: %t\n", plan.WillCreateDirs)
			fmt.Printf("  create_file: %t\n", plan.WillCreateFile)
			if plan.WillApplyTemplate {
				fmt.Printf("  template: %s\n", plan.TemplateAbs)
			}
			if strings.TrimSpace(content) == "" {
				fmt.Println("  content: (none supplied; would prompt interactively without --dry-run)")
			} else {
				fmt.Printf("  append_bytes: %d\n", len([]byte(content)))
			}
			return nil
		}

		return actions.AppendToDailyNote(&vault, content)
	},
}

func init() {
	appendCmd.Flags().BoolVarP(&appendTimestamp, "timestamp", "t", false, "prepend a timestamp to the content")
	appendCmd.Flags().StringVar(&appendTimeFmt, "time-format", "", "custom timestamp format (Go time format, default: 15:04)")
	appendCmd.Flags().BoolVar(&appendDryRun, "dry-run", false, "preview which note would be written without writing")
	appendCmd.Flags().StringVarP(&vaultName, "vault", "v", "", "vault name (not required if default is set)")
	rootCmd.AddCommand(appendCmd)
}

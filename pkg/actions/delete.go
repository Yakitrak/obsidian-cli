package actions

import (
	"bufio"
	"fmt"
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"os"
	"path/filepath"
	"strings"
)

type DeleteParams struct {
	NotePath string
	Force    bool
}

type DeletePlan struct {
	VaultName        string
	VaultPath        string
	RelativeNotePath string
	AbsoluteNotePath string
}

func PlanDeleteNote(vault obsidian.VaultManager, params DeleteParams) (DeletePlan, error) {
	vaultName, err := vault.DefaultName()
	if err != nil {
		return DeletePlan{}, err
	}

	vaultPath, err := vault.Path()
	if err != nil {
		return DeletePlan{}, err
	}

	rel := filepath.ToSlash(params.NotePath)
	abs, err := obsidian.SafeJoinVaultPath(vaultPath, rel)
	if err != nil {
		return DeletePlan{}, err
	}

	return DeletePlan{
		VaultName:        vaultName,
		VaultPath:        vaultPath,
		RelativeNotePath: rel,
		AbsoluteNotePath: abs,
	}, nil
}

func DeleteNote(vault obsidian.VaultManager, note obsidian.NoteManager, params DeleteParams) error {
	plan, err := PlanDeleteNote(vault, params)
	if err != nil {
		return err
	}

	if !params.Force {
		links, err := obsidian.FindIncomingLinks(plan.VaultPath, plan.RelativeNotePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not check for incoming links: %v\n", err)
		} else if len(links) > 0 {
			fmt.Printf("This note is linked from %d other note(s):\n", len(links))
			for _, link := range links {
				fmt.Printf("  - %s\n", link.SourcePath)
			}
			if !confirmDelete() {
				fmt.Println("Delete cancelled.")
				return nil
			}
		}
	}

	err = note.Delete(plan.AbsoluteNotePath)
	if err != nil {
		return err
	}
	return nil
}

func confirmDelete() bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Delete anyway? (y/N): ")
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}

package actions

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
)

// AppendToTarget appends content to the note resolved by the given target.
// If dryRun is true, it returns the computed plan without writing anything.
func AppendToTarget(vault *obsidian.Vault, targetName string, content string, now time.Time, dryRun bool) (obsidian.TargetPlan, error) {
	targetName = strings.TrimSpace(targetName)
	if targetName == "" {
		return obsidian.TargetPlan{}, errors.New("target name is required")
	}
	content = strings.TrimSpace(content)
	if content == "" {
		return obsidian.TargetPlan{}, errors.New("no content provided")
	}

	targets, err := obsidian.LoadTargets()
	if err != nil {
		return obsidian.TargetPlan{}, err
	}
	target, ok := targets[targetName]
	if !ok {
		return obsidian.TargetPlan{}, fmt.Errorf("target not found: %s", targetName)
	}

	effectiveVault := vault
	if strings.TrimSpace(target.Vault) != "" {
		effectiveVault = &obsidian.Vault{Name: strings.TrimSpace(target.Vault)}
	}

	vaultName, err := effectiveVault.DefaultName()
	if err != nil {
		return obsidian.TargetPlan{}, err
	}
	vaultPath, err := effectiveVault.Path()
	if err != nil {
		return obsidian.TargetPlan{}, err
	}

	plan, err := obsidian.PlanTargetAppend(vaultPath, targetName, target, now)
	if err != nil {
		return obsidian.TargetPlan{}, err
	}
	plan.VaultName = vaultName
	plan.VaultPath = vaultPath

	if dryRun {
		return plan, nil
	}

	if err := os.MkdirAll(filepath.Dir(plan.AbsoluteNotePath), 0750); err != nil {
		return obsidian.TargetPlan{}, fmt.Errorf("failed to create note directory: %w", err)
	}

	var existing []byte
	mode := os.FileMode(0600)
	if info, err := os.Stat(plan.AbsoluteNotePath); err == nil {
		mode = info.Mode()
		b, err := os.ReadFile(plan.AbsoluteNotePath)
		if err != nil {
			return obsidian.TargetPlan{}, fmt.Errorf("failed to read note: %w", err)
		}
		existing = b
	} else if err != nil && !os.IsNotExist(err) {
		return obsidian.TargetPlan{}, fmt.Errorf("failed to stat note: %w", err)
	} else {
		existing = []byte{}
		if plan.WillApplyTemplate {
			templateContent, err := os.ReadFile(plan.AbsoluteTemplate)
			if err != nil {
				return obsidian.TargetPlan{}, fmt.Errorf("failed to read template: %w", err)
			}
			title := filepath.Base(plan.AbsoluteNotePath)
			templateContent = obsidian.ExpandTemplateVariablesAt(templateContent, title, now)
			existing = templateContent
		}
	}

	next := appendWithSeparator(string(existing), content)
	if err := os.WriteFile(plan.AbsoluteNotePath, []byte(next), mode); err != nil {
		return obsidian.TargetPlan{}, fmt.Errorf("failed to write note: %w", err)
	}

	return plan, nil
}

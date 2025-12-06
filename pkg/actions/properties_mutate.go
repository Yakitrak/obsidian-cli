package actions

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/atomicobject/obsidian-cli/pkg/obsidian"
)

// PropertyMutationSummary captures the result of a property mutation operation.
type PropertyMutationSummary struct {
	NotesTouched    int            `json:"notesTouched"`
	PropertyChanges map[string]int `json:"propertyChanges"`
	FilesChanged    []string       `json:"filesChanged,omitempty"`
}

type propertyDelta struct {
	notesTouched bool
	changes      map[string]int
	fileChanged  string
	err          error
}

type propertyProcessor func(ctx context.Context, vaultPath, notePath string) propertyDelta

func runPropertyInParallelWithWorkers(ctx context.Context, cancel context.CancelFunc, vaultPath string, notes []string, processor propertyProcessor, workers int) (PropertyMutationSummary, error) {
	if workers < 1 {
		workers = 1
	}

	jobs := make(chan string, len(notes))
	results := make(chan propertyDelta, len(notes))

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for notePath := range jobs {
				select {
				case <-ctx.Done():
					return
				case results <- processor(ctx, vaultPath, notePath):
				}
			}
		}()
	}

	go func() {
		defer close(jobs)
		for _, notePath := range notes {
			select {
			case <-ctx.Done():
				return
			case jobs <- notePath:
			}
		}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	summary := PropertyMutationSummary{
		PropertyChanges: make(map[string]int),
		FilesChanged:    make([]string, 0),
	}

	var firstErr error
	var once sync.Once

	for delta := range results {
		if delta.err != nil {
			once.Do(func() {
				firstErr = delta.err
				cancel()
			})
			continue
		}

		if delta.notesTouched {
			summary.NotesTouched++
		}

		for prop, count := range delta.changes {
			summary.PropertyChanges[prop] += count
		}

		if delta.fileChanged != "" {
			summary.FilesChanged = append(summary.FilesChanged, delta.fileChanged)
		}
	}

	return summary, firstErr
}

func resolveVaultPath(vault obsidian.VaultManager) (string, error) {
	vaultPath, err := vault.Path()
	if err != nil {
		if v, ok := vault.(*obsidian.Vault); ok {
			if stat, statErr := os.Stat(v.Name); statErr == nil && stat.IsDir() {
				return v.Name, nil
			}
		}
		return "", fmt.Errorf("failed to get vault path: %w", err)
	}
	return vaultPath, nil
}

func findExistingKey(fm map[string]interface{}, target string) string {
	for key := range fm {
		if strings.EqualFold(key, target) {
			return key
		}
	}
	return ""
}

func processSetProperty(property string, value interface{}, overwrite, dryRun bool) propertyProcessor {
	return func(ctx context.Context, vaultPath, notePath string) propertyDelta {
		select {
		case <-ctx.Done():
			return propertyDelta{err: ctx.Err()}
		default:
		}

		full := filepath.Join(vaultPath, notePath)
		data, err := os.ReadFile(full)
		if err != nil {
			return propertyDelta{}
		}

		newContent, changed, err := obsidian.SetFrontmatterProperty(string(data), property, value, overwrite)
		if err != nil || !changed {
			return propertyDelta{err: err}
		}

		if !dryRun {
			if err := obsidian.WriteFileAtomic(full, []byte(newContent), 0644); err != nil {
				return propertyDelta{err: fmt.Errorf("failed to write file %s: %w", notePath, err)}
			}
		}

		return propertyDelta{
			notesTouched: true,
			changes:      map[string]int{property: 1},
			fileChanged:  notePath,
		}
	}
}

func processDeleteProperties(properties []string, dryRun bool) propertyProcessor {
	return func(ctx context.Context, vaultPath, notePath string) propertyDelta {
		select {
		case <-ctx.Done():
			return propertyDelta{err: ctx.Err()}
		default:
		}

		full := filepath.Join(vaultPath, notePath)
		data, err := os.ReadFile(full)
		if err != nil {
			return propertyDelta{}
		}

		content := string(data)
		fm, _ := obsidian.ExtractFrontmatter(content)
		var presentProps []string
		if fm != nil {
			for _, p := range properties {
				if key := findExistingKey(fm, p); key != "" {
					presentProps = append(presentProps, p)
				}
			}
		}

		newContent, changed, err := obsidian.DeleteFrontmatterProperties(content, properties)
		if err != nil || !changed {
			return propertyDelta{err: err}
		}

		if !dryRun {
			if err := obsidian.WriteFileAtomic(full, []byte(newContent), 0644); err != nil {
				return propertyDelta{err: fmt.Errorf("failed to write file %s: %w", notePath, err)}
			}
		}

		changes := make(map[string]int, len(presentProps))
		for _, p := range presentProps {
			changes[p] = 1
		}

		return propertyDelta{
			notesTouched: true,
			changes:      changes,
			fileChanged:  notePath,
		}
	}
}

func processRenameProperties(from []string, to string, merge, dryRun bool) propertyProcessor {
	return func(ctx context.Context, vaultPath, notePath string) propertyDelta {
		select {
		case <-ctx.Done():
			return propertyDelta{err: ctx.Err()}
		default:
		}

		full := filepath.Join(vaultPath, notePath)
		data, err := os.ReadFile(full)
		if err != nil {
			return propertyDelta{}
		}

		content := string(data)
		fm, _ := obsidian.ExtractFrontmatter(content)
		var presentSources []string
		destExists := false
		if fm != nil {
			for key := range fm {
				if strings.EqualFold(key, to) {
					destExists = true
				}
			}
			for _, p := range from {
				if key := findExistingKey(fm, p); key != "" {
					presentSources = append(presentSources, p)
				}
			}
		}

		newContent, changed, err := obsidian.RenameFrontmatterProperties(content, from, to, merge)
		if err != nil || !changed {
			return propertyDelta{err: err}
		}

		if !dryRun {
			if err := obsidian.WriteFileAtomic(full, []byte(newContent), 0644); err != nil {
				return propertyDelta{err: fmt.Errorf("failed to write file %s: %w", notePath, err)}
			}
		}

		changes := make(map[string]int, len(from)+1)
		for _, p := range presentSources {
			changes[p] = 1
		}
		destChanged := !destExists || merge
		if destChanged {
			changes[to]++
		}

		return propertyDelta{
			notesTouched: true,
			changes:      changes,
			fileChanged:  notePath,
		}
	}
}

// SetPropertyOnFiles sets a property on the provided files. files must be non-empty (pre-filtered by caller).
func SetPropertyOnFiles(vault obsidian.VaultManager, note obsidian.NoteManager, property string, value interface{}, files []string, overwrite, dryRun bool) (PropertyMutationSummary, error) {
	return SetPropertyOnFilesWithWorkers(vault, note, property, value, files, overwrite, dryRun, runtime.NumCPU())
}

// SetPropertyOnFilesWithWorkers sets a property on the provided files using workerCount workers.
func SetPropertyOnFilesWithWorkers(vault obsidian.VaultManager, note obsidian.NoteManager, property string, value interface{}, files []string, overwrite, dryRun bool, workerCount int) (PropertyMutationSummary, error) {
	if len(files) == 0 {
		return PropertyMutationSummary{}, fmt.Errorf("no files specified")
	}

	vaultPath, err := resolveVaultPath(vault)
	if err != nil {
		return PropertyMutationSummary{}, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	return runPropertyInParallelWithWorkers(ctx, cancel, vaultPath, files, processSetProperty(property, value, overwrite, dryRun), workerCount)
}

// DeleteProperties removes properties across the vault or provided files.
func DeleteProperties(vault obsidian.VaultManager, note obsidian.NoteManager, properties []string, files []string, dryRun bool) (PropertyMutationSummary, error) {
	return DeletePropertiesWithWorkers(vault, note, properties, files, dryRun, runtime.NumCPU())
}

func DeletePropertiesWithWorkers(vault obsidian.VaultManager, note obsidian.NoteManager, properties []string, files []string, dryRun bool, workerCount int) (PropertyMutationSummary, error) {
	if len(properties) == 0 {
		return PropertyMutationSummary{}, fmt.Errorf("no properties specified for deletion")
	}

	vaultPath, err := resolveVaultPath(vault)
	if err != nil {
		return PropertyMutationSummary{}, err
	}

	targetFiles := files
	if len(targetFiles) == 0 {
		targetFiles, err = note.GetNotesList(vaultPath)
		if err != nil {
			return PropertyMutationSummary{}, fmt.Errorf("failed to get notes list: %w", err)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	return runPropertyInParallelWithWorkers(ctx, cancel, vaultPath, targetFiles, processDeleteProperties(properties, dryRun), workerCount)
}

// RenameProperties renames one or more properties to a single destination across the vault or provided files.
func RenameProperties(vault obsidian.VaultManager, note obsidian.NoteManager, from []string, to string, merge bool, files []string, dryRun bool) (PropertyMutationSummary, error) {
	return RenamePropertiesWithWorkers(vault, note, from, to, merge, files, dryRun, runtime.NumCPU())
}

func RenamePropertiesWithWorkers(vault obsidian.VaultManager, note obsidian.NoteManager, from []string, to string, merge bool, files []string, dryRun bool, workerCount int) (PropertyMutationSummary, error) {
	if len(from) == 0 {
		return PropertyMutationSummary{}, fmt.Errorf("no source properties specified")
	}
	if to == "" {
		return PropertyMutationSummary{}, fmt.Errorf("destination property cannot be empty")
	}

	vaultPath, err := resolveVaultPath(vault)
	if err != nil {
		return PropertyMutationSummary{}, err
	}

	targetFiles := files
	if len(targetFiles) == 0 {
		targetFiles, err = note.GetNotesList(vaultPath)
		if err != nil {
			return PropertyMutationSummary{}, fmt.Errorf("failed to get notes list: %w", err)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	return runPropertyInParallelWithWorkers(ctx, cancel, vaultPath, targetFiles, processRenameProperties(from, to, merge, dryRun), workerCount)
}

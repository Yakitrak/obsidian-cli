package actions

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"github.com/stretchr/testify/assert"
)

func TestDeleteTagsIntegration(t *testing.T) {
	tests := []struct {
		name             string
		tagsToDelete     []string
		fileContent      string
		expectedChanged  bool
		shouldContain    []string
		shouldNotContain []string
	}{
		{
			name:         "delete frontmatter tag",
			tagsToDelete: []string{"work"},
			fileContent: `---
title: Test Note
tags: [work, personal]
---
# Test Note
Some content here.`,
			expectedChanged:  true,
			shouldContain:    []string{"personal"},
			shouldNotContain: []string{"work"},
		},
		{
			name:         "delete hashtag",
			tagsToDelete: []string{"work"},
			fileContent: `# Test Note
This is about #work and other things.
More content here.`,
			expectedChanged:  true,
			shouldNotContain: []string{"#work"},
		},
		{
			name:             "ignore hashtags in code blocks",
			tagsToDelete:     []string{"work"},
			fileContent:      "# Test Note\nThis is about #work.\n\n```\n#work should not be deleted here\n```\n\nMore #work content.",
			expectedChanged:  true,
			shouldContain:    []string{"#work should not be deleted here"},
			shouldNotContain: []string{"This is about #work", "More #work content"},
		},
		{
			name:         "no changes when tag not found",
			tagsToDelete: []string{"nonexistent"},
			fileContent: `---
tags: [personal]
---
# Test Note
This is about #other things.`,
			expectedChanged: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory
			tempDir, err := os.MkdirTemp("", "obsidian-test-*")
			assert.NoError(t, err)
			defer os.RemoveAll(tempDir)

			// Create test file
			testFile := filepath.Join(tempDir, "test.md")
			err = os.WriteFile(testFile, []byte(tt.fileContent), 0644)
			assert.NoError(t, err)

			// Create vault and note managers
			vault := &obsidian.Vault{Name: filepath.Base(tempDir)}
			note := &obsidian.Note{}

			// Mock vault path to return our temp dir
			vault.Name = tempDir // This is a hack but works for testing

			// Execute delete (not dry run)
			summary, err := DeleteTags(vault, note, tt.tagsToDelete, false)
			assert.NoError(t, err)

			if tt.expectedChanged {
				assert.Greater(t, summary.NotesTouched, 0)

				// Read the modified file
				modifiedContent, err := os.ReadFile(testFile)
				assert.NoError(t, err)
				contentStr := string(modifiedContent)

				// Check expected content
				for _, shouldContain := range tt.shouldContain {
					assert.Contains(t, contentStr, shouldContain, "Should contain: %s", shouldContain)
				}
				for _, shouldNotContain := range tt.shouldNotContain {
					assert.NotContains(t, contentStr, shouldNotContain, "Should not contain: %s", shouldNotContain)
				}
			} else {
				assert.Equal(t, 0, summary.NotesTouched)
			}
		})
	}
}

func TestRenameTagsIntegration(t *testing.T) {
	tests := []struct {
		name             string
		fromTags         []string
		toTag            string
		fileContent      string
		expectedChanged  bool
		shouldContain    []string
		shouldNotContain []string
	}{
		{
			name:     "rename frontmatter tag",
			fromTags: []string{"work"},
			toTag:    "office",
			fileContent: `---
title: Test Note
tags: [work, personal]
---
# Test Note
Some content here.`,
			expectedChanged:  true,
			shouldContain:    []string{"office", "personal"},
			shouldNotContain: []string{"work"},
		},
		{
			name:     "rename hashtag",
			fromTags: []string{"work"},
			toTag:    "office",
			fileContent: `# Test Note
This is about #work and other things.
More #work content here.`,
			expectedChanged:  true,
			shouldContain:    []string{"#office"},
			shouldNotContain: []string{"#work"},
		},
		{
			name:     "rename multiple tags to same destination",
			fromTags: []string{"work", "job"},
			toTag:    "office",
			fileContent: `---
tags: [work, personal, job]
---
# Test Note
Some content here.`,
			expectedChanged:  true,
			shouldContain:    []string{"office", "personal"},
			shouldNotContain: []string{"work", "job"},
		},
		{
			name:     "hierarchical rename in mixed content",
			fromTags: []string{"work"},
			toTag:    "project",
			fileContent: `---
tags: [work/frontend, work/backend, other]
---
# Test Note
Discussing #work/testing and #work/deployment but not #working.`,
			expectedChanged:  true,
			shouldContain:    []string{"project/frontend", "project/backend", "project/testing", "project/deployment", "#working"},
			shouldNotContain: []string{"work/frontend", "work/backend", "#work/testing", "#work/deployment"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory
			tempDir, err := os.MkdirTemp("", "obsidian-test-*")
			assert.NoError(t, err)
			defer os.RemoveAll(tempDir)

			// Create test file
			testFile := filepath.Join(tempDir, "test.md")
			err = os.WriteFile(testFile, []byte(tt.fileContent), 0644)
			assert.NoError(t, err)

			// Create vault and note managers
			vault := &obsidian.Vault{Name: tempDir} // Hack for testing
			note := &obsidian.Note{}

			// Execute rename (not dry run)
			summary, err := RenameTags(vault, note, tt.fromTags, tt.toTag, false)
			assert.NoError(t, err)

			if tt.expectedChanged {
				assert.Greater(t, summary.NotesTouched, 0)

				// Read the modified file
				modifiedContent, err := os.ReadFile(testFile)
				assert.NoError(t, err)
				contentStr := string(modifiedContent)

				// Check expected content
				for _, shouldContain := range tt.shouldContain {
					assert.Contains(t, contentStr, shouldContain, "Should contain: %s", shouldContain)
				}
				for _, shouldNotContain := range tt.shouldNotContain {
					assert.NotContains(t, contentStr, shouldNotContain, "Should not contain: %s", shouldNotContain)
				}
			} else {
				assert.Equal(t, 0, summary.NotesTouched)
			}
		})
	}
}

func TestDeleteTagsValidation(t *testing.T) {
	vault := &obsidian.Vault{Name: "/fake/path"}
	note := &obsidian.Note{}

	tests := []struct {
		name         string
		tagsToDelete []string
		expectError  bool
		errorMessage string
	}{
		{
			name:         "empty tags list",
			tagsToDelete: []string{},
			expectError:  true,
			errorMessage: "no tags specified for deletion",
		},
		{
			name:         "invalid tag with spaces",
			tagsToDelete: []string{"invalid tag"},
			expectError:  true,
			errorMessage: "invalid tag",
		},
		{
			name:         "purely numeric tag",
			tagsToDelete: []string{"123"},
			expectError:  true,
			errorMessage: "invalid tag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DeleteTags(vault, note, tt.tagsToDelete, true)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMessage)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRenameTagsValidation(t *testing.T) {
	vault := &obsidian.Vault{Name: "/fake/path"}
	note := &obsidian.Note{}

	tests := []struct {
		name         string
		fromTags     []string
		toTag        string
		expectError  bool
		errorMessage string
	}{
		{
			name:         "empty from tags",
			fromTags:     []string{},
			toTag:        "office",
			expectError:  true,
			errorMessage: "no source tags specified for rename",
		},
		{
			name:         "empty to tag",
			fromTags:     []string{"work"},
			toTag:        "",
			expectError:  true,
			errorMessage: "destination tag cannot be empty",
		},
		{
			name:         "invalid source tag",
			fromTags:     []string{"invalid tag"},
			toTag:        "office",
			expectError:  true,
			errorMessage: "invalid source tag",
		},
		{
			name:         "invalid destination tag",
			fromTags:     []string{"work"},
			toTag:        "invalid tag",
			expectError:  true,
			errorMessage: "invalid destination tag",
		},
		{
			name:         "circular rename",
			fromTags:     []string{"work", "office"},
			toTag:        "work",
			expectError:  true,
			errorMessage: "cannot rename tag work to itself",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := RenameTags(vault, note, tt.fromTags, tt.toTag, true)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMessage)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDryRunDoesNotModifyFiles(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "obsidian-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test file
	originalContent := `---
tags: [work, personal]
---
# Test Note
This is about #work.`

	testFile := filepath.Join(tempDir, "test.md")
	err = os.WriteFile(testFile, []byte(originalContent), 0644)
	assert.NoError(t, err)

	// Create vault and note managers
	vault := &obsidian.Vault{Name: tempDir} // Hack for testing
	note := &obsidian.Note{}

	// Execute dry run delete
	summary, err := DeleteTags(vault, note, []string{"work"}, true)
	assert.NoError(t, err)
	assert.Greater(t, summary.NotesTouched, 0)

	// Check file was not modified
	actualContent, err := os.ReadFile(testFile)
	assert.NoError(t, err)
	assert.Equal(t, originalContent, string(actualContent))
}

func TestAddTagsIntegration(t *testing.T) {
	tests := []struct {
		name            string
		tagsToAdd       []string
		fileContent     string
		expectedChanged bool
		shouldContain   []string
	}{
		{
			name:      "add tags to existing frontmatter",
			tagsToAdd: []string{"urgent", "project"},
			fileContent: `---
title: Test Note
tags: [work]
---
# Test Note
Some content here.`,
			expectedChanged: true,
			shouldContain:   []string{"work", "urgent", "project"},
		},
		{
			name:      "add tags to note without frontmatter",
			tagsToAdd: []string{"new-tag"},
			fileContent: `# Test Note
This is a note without frontmatter.`,
			expectedChanged: true,
			shouldContain:   []string{"new-tag"},
		},
		{
			name:      "avoid duplicate tags",
			tagsToAdd: []string{"work", "duplicate"},
			fileContent: `---
title: Test Note
tags: [work, existing]
---
# Test Note
Some content here.`,
			expectedChanged: true,
			shouldContain:   []string{"work", "existing", "duplicate"},
		},
		{
			name:      "no change when all tags already exist",
			tagsToAdd: []string{"work", "existing"},
			fileContent: `---
title: Test Note
tags: [work, existing]
---
# Test Note
Some content here.`,
			expectedChanged: false,
			shouldContain:   []string{"work", "existing"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory and file
			tempDir, err := os.MkdirTemp("", "add_tags_test")
			assert.NoError(t, err)
			defer os.RemoveAll(tempDir)

			testFile := filepath.Join(tempDir, "test.md")
			err = os.WriteFile(testFile, []byte(tt.fileContent), 0644)
			assert.NoError(t, err)

			// Mock vault and note
			vault := &obsidian.Vault{Name: tempDir}
			note := &obsidian.Note{}

			// Execute add tags
			summary, err := AddTagsToFiles(vault, note, tt.tagsToAdd, []string{"test.md"}, false)
			assert.NoError(t, err)

			if tt.expectedChanged {
				assert.Equal(t, 1, summary.NotesTouched)
				assert.Greater(t, len(summary.TagChanges), 0)
			} else {
				assert.Equal(t, 0, summary.NotesTouched)
			}

			// Check file content
			actualContent, err := os.ReadFile(testFile)
			assert.NoError(t, err)
			actualStr := string(actualContent)

			for _, shouldContain := range tt.shouldContain {
				assert.Contains(t, actualStr, shouldContain, "Should contain tag: %s", shouldContain)
			}
		})
	}
}

func TestAddTagsWithEmptyTags(t *testing.T) {
	vault := &obsidian.Vault{Name: "/tmp"}
	note := &obsidian.Note{}

	_, err := AddTagsToFiles(vault, note, []string{}, []string{"test.md"}, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no tags specified")
}

func TestAddTagsWithEmptyFiles(t *testing.T) {
	vault := &obsidian.Vault{Name: "/tmp"}
	note := &obsidian.Note{}

	_, err := AddTagsToFiles(vault, note, []string{"test"}, []string{}, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no files specified")
}

func TestAddTagsDryRun(t *testing.T) {
	// Create temp directory and file
	tempDir, err := os.MkdirTemp("", "add_tags_dry_run_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	originalContent := `---
title: Test Note
tags: [existing]
---
# Test Note
Some content here.`

	testFile := filepath.Join(tempDir, "test.md")
	err = os.WriteFile(testFile, []byte(originalContent), 0644)
	assert.NoError(t, err)

	// Mock vault and note
	vault := &obsidian.Vault{Name: tempDir}
	note := &obsidian.Note{}

	// Execute dry run add
	summary, err := AddTagsToFiles(vault, note, []string{"new-tag"}, []string{"test.md"}, true)
	assert.NoError(t, err)
	assert.Greater(t, summary.NotesTouched, 0)

	// Check file was not modified
	actualContent, err := os.ReadFile(testFile)
	assert.NoError(t, err)
	assert.Equal(t, originalContent, string(actualContent))
}

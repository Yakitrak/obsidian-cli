package actions_test

import (
	"errors"
	"testing"

	"github.com/Yakitrak/obsidian-cli/mocks"
	"github.com/Yakitrak/obsidian-cli/pkg/actions"
	"github.com/stretchr/testify/assert"
)

func TestFrontmatter_Print(t *testing.T) {
	t.Run("Print frontmatter successfully", func(t *testing.T) {
		vault := mocks.MockVaultOperator{Name: "myVault"}
		note := mocks.MockNoteManager{
			Contents: "---\ntitle: Test Note\ntags:\n  - a\n  - b\n---\nBody content",
		}

		output, err := actions.Frontmatter(&vault, &note, actions.FrontmatterParams{
			NoteName: "test-note",
			Print:    true,
		})

		assert.NoError(t, err)
		assert.Contains(t, output, "title: Test Note")
		assert.Contains(t, output, "tags:")
	})

	t.Run("Print empty for note without frontmatter", func(t *testing.T) {
		vault := mocks.MockVaultOperator{Name: "myVault"}
		note := mocks.MockNoteManager{
			Contents: "Just body content without frontmatter",
		}

		output, err := actions.Frontmatter(&vault, &note, actions.FrontmatterParams{
			NoteName: "test-note",
			Print:    true,
		})

		assert.NoError(t, err)
		assert.Empty(t, output)
	})

	t.Run("Vault error propagates", func(t *testing.T) {
		vault := mocks.MockVaultOperator{
			DefaultNameErr: errors.New("vault error"),
		}
		note := mocks.MockNoteManager{}

		_, err := actions.Frontmatter(&vault, &note, actions.FrontmatterParams{
			NoteName: "test-note",
			Print:    true,
		})

		assert.Error(t, err)
	})

	t.Run("Note error propagates", func(t *testing.T) {
		vault := mocks.MockVaultOperator{Name: "myVault"}
		note := mocks.MockNoteManager{
			GetContentsError: errors.New("note not found"),
		}

		_, err := actions.Frontmatter(&vault, &note, actions.FrontmatterParams{
			NoteName: "test-note",
			Print:    true,
		})

		assert.Error(t, err)
	})
}

func TestFrontmatter_Edit(t *testing.T) {
	t.Run("Edit existing frontmatter key", func(t *testing.T) {
		vault := mocks.MockVaultOperator{Name: "myVault"}
		note := mocks.MockNoteManager{
			Contents: "---\ntitle: Old Title\n---\nBody",
		}

		output, err := actions.Frontmatter(&vault, &note, actions.FrontmatterParams{
			NoteName: "test-note",
			Edit:     true,
			Key:      "title",
			Value:    "New Title",
		})

		assert.NoError(t, err)
		assert.Contains(t, output, "Updated frontmatter key 'title'")
	})

	t.Run("Add new frontmatter key", func(t *testing.T) {
		vault := mocks.MockVaultOperator{Name: "myVault"}
		note := mocks.MockNoteManager{
			Contents: "---\ntitle: Test\n---\nBody",
		}

		output, err := actions.Frontmatter(&vault, &note, actions.FrontmatterParams{
			NoteName: "test-note",
			Edit:     true,
			Key:      "author",
			Value:    "John",
		})

		assert.NoError(t, err)
		assert.Contains(t, output, "Updated frontmatter key 'author'")
	})

	t.Run("Create frontmatter when none exists", func(t *testing.T) {
		vault := mocks.MockVaultOperator{Name: "myVault"}
		note := mocks.MockNoteManager{
			Contents: "Just body content",
		}

		output, err := actions.Frontmatter(&vault, &note, actions.FrontmatterParams{
			NoteName: "test-note",
			Edit:     true,
			Key:      "title",
			Value:    "New Note",
		})

		assert.NoError(t, err)
		assert.Contains(t, output, "Updated frontmatter key 'title'")
	})

	t.Run("Edit requires key", func(t *testing.T) {
		vault := mocks.MockVaultOperator{Name: "myVault"}
		note := mocks.MockNoteManager{
			Contents: "---\ntitle: Test\n---\nBody",
		}

		_, err := actions.Frontmatter(&vault, &note, actions.FrontmatterParams{
			NoteName: "test-note",
			Edit:     true,
			Value:    "value",
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "--key is required")
	})

	t.Run("Edit requires value", func(t *testing.T) {
		vault := mocks.MockVaultOperator{Name: "myVault"}
		note := mocks.MockNoteManager{
			Contents: "---\ntitle: Test\n---\nBody",
		}

		_, err := actions.Frontmatter(&vault, &note, actions.FrontmatterParams{
			NoteName: "test-note",
			Edit:     true,
			Key:      "title",
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "--value is required")
	})

	t.Run("Write error propagates", func(t *testing.T) {
		vault := mocks.MockVaultOperator{Name: "myVault"}
		note := mocks.MockNoteManager{
			Contents:         "---\ntitle: Test\n---\nBody",
			SetContentsError: errors.New("write error"),
		}

		_, err := actions.Frontmatter(&vault, &note, actions.FrontmatterParams{
			NoteName: "test-note",
			Edit:     true,
			Key:      "title",
			Value:    "New",
		})

		assert.Error(t, err)
	})
}

func TestFrontmatter_Delete(t *testing.T) {
	t.Run("Delete existing key", func(t *testing.T) {
		vault := mocks.MockVaultOperator{Name: "myVault"}
		note := mocks.MockNoteManager{
			Contents: "---\ntitle: Test\nauthor: John\n---\nBody",
		}

		output, err := actions.Frontmatter(&vault, &note, actions.FrontmatterParams{
			NoteName: "test-note",
			Delete:   true,
			Key:      "author",
		})

		assert.NoError(t, err)
		assert.Contains(t, output, "Deleted frontmatter key 'author'")
	})

	t.Run("Delete requires key", func(t *testing.T) {
		vault := mocks.MockVaultOperator{Name: "myVault"}
		note := mocks.MockNoteManager{
			Contents: "---\ntitle: Test\n---\nBody",
		}

		_, err := actions.Frontmatter(&vault, &note, actions.FrontmatterParams{
			NoteName: "test-note",
			Delete:   true,
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "--key is required")
	})

	t.Run("Delete from note without frontmatter returns error", func(t *testing.T) {
		vault := mocks.MockVaultOperator{Name: "myVault"}
		note := mocks.MockNoteManager{
			Contents: "Just body content",
		}

		_, err := actions.Frontmatter(&vault, &note, actions.FrontmatterParams{
			NoteName: "test-note",
			Delete:   true,
			Key:      "title",
		})

		assert.Error(t, err)
	})
}

func TestFrontmatter_NoOperation(t *testing.T) {
	t.Run("Returns error when no operation specified", func(t *testing.T) {
		vault := mocks.MockVaultOperator{Name: "myVault"}
		note := mocks.MockNoteManager{
			Contents: "---\ntitle: Test\n---\nBody",
		}

		_, err := actions.Frontmatter(&vault, &note, actions.FrontmatterParams{
			NoteName: "test-note",
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no operation specified")
	})
}

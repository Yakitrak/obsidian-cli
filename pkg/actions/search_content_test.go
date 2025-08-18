package actions_test

import (
	"errors"
	"testing"

	"github.com/Yakitrak/obsidian-cli/mocks"
	"github.com/Yakitrak/obsidian-cli/pkg/actions"
	"github.com/stretchr/testify/assert"
)

func TestSearchNotesContent(t *testing.T) {
	t.Run("Successful content search with single match", func(t *testing.T) {
		vault := mocks.MockVaultOperator{Name: "myVault"}
		uri := mocks.MockUriManager{}
		note := mocks.MockNoteManager{}
		fuzzyFinder := mocks.MockFuzzyFinder{}

		err := actions.SearchNotesContent(&vault, &note, &uri, &fuzzyFinder, "test")
		assert.NoError(t, err)
	})

	t.Run("No matches found", func(t *testing.T) {
		vault := mocks.MockVaultOperator{Name: "myVault"}
		uri := mocks.MockUriManager{}
		note := mocks.MockNoteManager{NoMatches: true}
		fuzzyFinder := mocks.MockFuzzyFinder{}

		err := actions.SearchNotesContent(&vault, &note, &uri, &fuzzyFinder, "nonexistent")
		assert.NoError(t, err)
	})

	t.Run("SearchNotesWithSnippets returns error", func(t *testing.T) {
		vault := mocks.MockVaultOperator{Name: "myVault"}
		uri := mocks.MockUriManager{}
		note := mocks.MockNoteManager{
			GetContentsError: errors.New("search failed"),
		}
		fuzzyFinder := mocks.MockFuzzyFinder{}

		err := actions.SearchNotesContent(&vault, &note, &uri, &fuzzyFinder, "test")
		assert.Error(t, err)
	})

	t.Run("vault.DefaultName returns error", func(t *testing.T) {
		vault := mocks.MockVaultOperator{
			DefaultNameErr: errors.New("vault name error"),
		}
		uri := mocks.MockUriManager{}
		note := mocks.MockNoteManager{}
		fuzzyFinder := mocks.MockFuzzyFinder{}

		err := actions.SearchNotesContent(&vault, &note, &uri, &fuzzyFinder, "test")
		assert.Error(t, err)
	})

	t.Run("vault.Path returns error", func(t *testing.T) {
		vault := mocks.MockVaultOperator{
			PathError: errors.New("vault path error"),
		}
		uri := mocks.MockUriManager{}
		note := mocks.MockNoteManager{}
		fuzzyFinder := mocks.MockFuzzyFinder{}

		err := actions.SearchNotesContent(&vault, &note, &uri, &fuzzyFinder, "test")
		assert.Error(t, err)
	})

	t.Run("fuzzy finder returns error", func(t *testing.T) {
		vault := mocks.MockVaultOperator{Name: "myVault"}
		uri := mocks.MockUriManager{}
		note := mocks.MockNoteManager{}
		fuzzyFinder := mocks.MockFuzzyFinder{
			FindErr: errors.New("fuzzy finder error"),
		}

		err := actions.SearchNotesContent(&vault, &note, &uri, &fuzzyFinder, "test")
		assert.Error(t, err)
	})

	t.Run("uri execution returns error", func(t *testing.T) {
		vault := mocks.MockVaultOperator{Name: "myVault"}
		uri := mocks.MockUriManager{
			ExecuteErr: errors.New("uri execution error"),
		}
		note := mocks.MockNoteManager{}
		fuzzyFinder := mocks.MockFuzzyFinder{}

		err := actions.SearchNotesContent(&vault, &note, &uri, &fuzzyFinder, "test")
		assert.Error(t, err)
	})
}

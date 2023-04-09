package pkg_test

import (
	"github.com/Yakitrak/obsidian-cli/pkg"
	"testing"
)

func TestMoveNote(t *testing.T) {
	var calledBaseUri string
	var calledVaultName string
	var mockVaultPath = "vault-path"
	var calledCurrentPath string
	var calledNewPath string
	var calledCurrentNoteName string
	var calledNewNoteName string

	mockUriConstructor := func(baseUri string, params map[string]string) string {
		calledBaseUri = baseUri
		calledVaultName = params["vault"]
		return ""
	}

	mockFindVaultPathFromName := func(vaultName string) (string, error) {
		return mockVaultPath, nil
	}

	mockMoveNote := func(currentPath string, newPath string) {
		calledCurrentPath = currentPath
		calledNewPath = newPath
		return
	}

	mockUpdateLinksInVault := func(vaultPath string, currentNoteName string, newNoteName string) {
		calledCurrentNoteName = currentNoteName
		calledNewNoteName = newNoteName
		return
	}

	var tests = []struct {
		testName        string
		vaultName       string
		currentNoteName string
		newNoteName     string
	}{

		{"Given vault", "v-name", "current-note-name", "new-note-name"},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {

			pkg.MoveNote(mockUriConstructor, mockFindVaultPathFromName, mockMoveNote, mockUpdateLinksInVault, tt.vaultName, tt.currentNoteName, tt.newNoteName)

			t.Run("mockUriConstructor call", func(t *testing.T) {
				if calledBaseUri != pkg.ObsOpenUrl {
					t.Errorf("got %s, want %s", calledBaseUri, pkg.ObsSearchUrl)
				}
				if calledVaultName != tt.vaultName {
					t.Errorf("got %s, want %s", calledVaultName, tt.vaultName)
				}
			})

			t.Run("mockMoveNote call", func(t *testing.T) {
				if calledCurrentPath != mockVaultPath+"/"+tt.currentNoteName {
					t.Errorf("got %s, want %s", calledCurrentPath, mockVaultPath+"/"+tt.currentNoteName)
				}
				if calledNewPath != mockVaultPath+"/"+tt.newNoteName {
					t.Errorf("got %s, want %s", calledNewPath, mockVaultPath+"/"+tt.newNoteName)
				}
			})

			t.Run("mockUpdateLinksInVault call", func(t *testing.T) {
				if calledCurrentNoteName != tt.currentNoteName {
					t.Errorf("got %s, want %s", calledCurrentNoteName, tt.currentNoteName)
				}
				if calledNewNoteName != tt.newNoteName {
					t.Errorf("got %s, want %s", calledNewNoteName, tt.newNoteName)
				}
			})
		})

	}

}

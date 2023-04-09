package pkg_test

import (
	"github.com/Yakitrak/obsidian-cli/pkg"
	"testing"
)

func TestOpenNote(t *testing.T) {
	var calledBaseUri string
	var calledVaultName string
	var calledNoteName string

	mockUriConstructor := func(baseUri string, params map[string]string) string {
		calledBaseUri = baseUri
		calledVaultName = params["vault"]
		calledNoteName = params["file"]
		return ""
	}

	var tests = []struct {
		testName  string
		vaultName string
		noteName  string
	}{
		{"Given direct file", "v-name", "name"},
		{"Given Nested path", "v-name", "nested path/another path/file here"},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			pkg.OpenNote(mockUriConstructor, tt.vaultName, tt.noteName)

			t.Run("Then it should call the uri constructor with the correct parameters", func(t *testing.T) {
				if calledBaseUri != pkg.ObsOpenUrl {
					t.Errorf("got %s, want %s", calledBaseUri, pkg.ObsOpenUrl)
				}
				if calledVaultName != tt.vaultName {
					t.Errorf("got %s, want %s", calledVaultName, tt.vaultName)
				}
				if calledNoteName != tt.noteName {
					t.Errorf("got %s, want %s", calledNoteName, tt.noteName)
				}
			})
		})
	}
}

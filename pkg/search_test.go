package pkg_test

import (
	"github.com/Yakitrak/obsidian-cli/pkg"
	"testing"
)

func TestSearchNotes(t *testing.T) {
	var calledBaseUri string
	var calledVaultName string
	var calledSearchText string

	mockUriConstructor := func(baseUri string, params map[string]string) string {
		calledBaseUri = baseUri
		calledVaultName = params["vault"]
		calledSearchText = params["query"]
		return ""
	}

	var tests = []struct {
		testName   string
		vaultName  string
		searchText string
	}{

		{"Given vault", "v-name", "search text"},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			pkg.SearchNotes(mockUriConstructor, tt.vaultName, tt.searchText)
			t.Run("Then it should call the uri constructor with the correct parameters", func(t *testing.T) {
				if calledBaseUri != pkg.ObsSearchUrl {
					t.Errorf("got %s, want %s", calledBaseUri, pkg.ObsSearchUrl)
				}
				if calledVaultName != tt.vaultName {
					t.Errorf("got %s, want %s", calledVaultName, tt.vaultName)
				}
				if calledSearchText != tt.searchText {
					t.Errorf("got %s, want %s", calledSearchText, tt.searchText)
				}
			})
		})
	}
}

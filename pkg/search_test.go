package pkg

import (
	"testing"

	"github.com/Yakitrak/obsidian-cli/utils"
)

func TestSearchNotes(t *testing.T) {
	// Set up default vault
	utils.SetDefaultVault("default-v")

	// Defining the columns of the table
	var tests = []struct {
		name       string
		searchText string
		vaultName  string
		want       string
	}{

		{"Given vault", "search text", "v-name", ObsSearchUrl + "?query=search%20text&vault=v-name"},
		{"No vault", "search text", "", ObsSearchUrl + "?query=search%20text&vault=default-v"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SearchNotes(tt.searchText, tt.vaultName)
			if utils.MapsEqual(utils.ExtractParams(got), utils.ExtractParams(tt.want)) == false {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}

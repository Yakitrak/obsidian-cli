package pkg_test

import (
	"github.com/Yakitrak/obsidian-cli/pkg"
	"testing"

	"github.com/Yakitrak/obsidian-cli/utils"
)

func TestSearchNotes(t *testing.T) {
	// Defining the columns of the table
	var tests = []struct {
		name       string
		searchText string
		vaultName  string
		want       string
	}{

		{"Given vault", "search text", "v-name", pkg.ObsSearchUrl + "?query=search%20text&vault=v-name"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pkg.SearchNotes(tt.searchText, tt.vaultName)
			if utils.MapsEqual(utils.ExtractParams(got), utils.ExtractParams(tt.want)) == false {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}

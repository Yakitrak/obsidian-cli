package pkg

import (
	"testing"

	"github.com/Yakitrak/obsidian-cli/utils"
)

func TestCreateNote(t *testing.T) {
	// Set up default vault
	utils.SetDefaultVault("default-v")

	var tests = []struct {
		testName        string
		vaultName       string
		noteName        string
		content         string
		shouldAppend    bool
		shouldOverwrite bool
		want            string
	}{

		{"Given vault with no content", "v-name", "note", "", false, false, ObsCreateUrl + "?file=note&vault=v-name"},
		{"No given vault with no content", "", "note", "", false, false, ObsCreateUrl + "?file=note&vault=default-v"},
		{"Given content", "", "note", "content", false, false, ObsCreateUrl + "?vault=default-v&content=content&file=note"},
		{"Given content to overwrite", "", "note", "content", false, true, ObsCreateUrl + "?content=content&file=note&overwrite=true&vault=default-v"},
		{"Given content to append", "", "note", "content", true, false, ObsCreateUrl + "?append=true&content=content&vault=default-v&file=note"},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got := CreateNote(tt.vaultName, tt.noteName, tt.content, tt.shouldAppend, tt.shouldOverwrite)
			if utils.MapsEqual(utils.ExtractParams(got), utils.ExtractParams(tt.want)) == false {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}

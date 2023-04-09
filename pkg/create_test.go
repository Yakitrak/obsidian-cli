package pkg_test

import (
	"github.com/Yakitrak/obsidian-cli/pkg"
	"testing"

	"github.com/Yakitrak/obsidian-cli/utils"
)

func TestCreateNote(t *testing.T) {
	var tests = []struct {
		testName        string
		vaultName       string
		noteName        string
		content         string
		shouldAppend    bool
		shouldOverwrite bool
		want            string
	}{

		{"Given no content", "v-name", "note", "", false, false, pkg.ObsCreateUrl + "?file=note&vault=v-name"},
		{"Given content", "default-v", "note", "content", false, false, pkg.ObsCreateUrl + "?vault=default-v&content=content&file=note"},
		{"Given content to overwrite", "default-v", "note", "content", false, true, pkg.ObsCreateUrl + "?content=content&file=note&overwrite=true&vault=default-v"},
		{"Given content to append", "default-v", "note", "content", true, false, pkg.ObsCreateUrl + "?append=true&content=content&vault=default-v&file=note"},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got := pkg.CreateNote(tt.vaultName, tt.noteName, tt.content, tt.shouldAppend, tt.shouldOverwrite)
			if utils.MapsEqual(utils.ExtractParams(got), utils.ExtractParams(tt.want)) == false {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}

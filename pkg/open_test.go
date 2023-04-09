package pkg_test

import (
	"github.com/Yakitrak/obsidian-cli/pkg"
	"testing"

	"github.com/Yakitrak/obsidian-cli/utils"
)

func TestOpenNote(t *testing.T) {

	// Defining the columns of the table
	var tests = []struct {
		name      string
		noteName  string
		vaultName string
		want      string
	}{
		{"Given direct file", "name", "v-name", pkg.ObsOpenUrl + "?file=name&vault=v-name"},
		{"Given Nested path", "nested path/another path/file here", "v-name", pkg.ObsOpenUrl + "?file=nested%20path%2Fanother%20path%2Ffile%20here&vault=v-name"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pkg.OpenNote(tt.noteName, tt.vaultName)
			if utils.MapsEqual(utils.ExtractParams(got), utils.ExtractParams(tt.want)) == false {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}

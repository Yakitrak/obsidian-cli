package pkg

import (
	"testing"

	"github.com/Yakitrak/obsidian-cli/utils"
)

func TestOpenNote(t *testing.T) {
	// Set up default vault
	utils.SetDefaultVault("default-v")

	// Defining the columns of the table
	var tests = []struct {
		name      string
		noteName  string
		vaultName string
		want      string
	}{
		{"Given vault", "name", "v-name", ObsOpenUrl + "?file=name&vault=v-name"},
		{"No vault", "name", "", ObsOpenUrl + "?file=name&vault=default-v"},
		{"Nested path", "nested path/another path/file here", "", ObsOpenUrl + "?file=nested%20path%2Fanother%20path%2Ffile%20here&vault=default-v"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := OpenNote(tt.noteName, tt.vaultName)
			if utils.MapsEqual(utils.ExtractParams(got), utils.ExtractParams(tt.want)) == false {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}

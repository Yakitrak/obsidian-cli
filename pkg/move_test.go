package pkg_test

import (
	"github.com/Yakitrak/obsidian-cli/pkg"
	"testing"

	"github.com/Yakitrak/obsidian-cli/utils"
)

func TestMoveNote(t *testing.T) {
	// Defining the columns of the table
	var tests = []struct {
		name            string
		vaultName       string
		currentNoteName string
		newNoteName     string
		want            string
	}{

		{"Given vault", "v-name", "current-name", "new-name", pkg.ObsOpenUrl + "?open=search%20text&vault=v-name"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pkg.MoveNote(tt.vaultName, tt.currentNoteName, tt.newNoteName)
			if utils.MapsEqual(utils.ExtractParams(got), utils.ExtractParams(tt.want)) == false {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}

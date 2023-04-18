package actions_test

import (
	"github.com/Yakitrak/obsidian-cli/pkg/actions"
	"github.com/stretchr/testify/assert"
	"testing"
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

		{"Given no content", "v-name", "note", "", false, false, actions.ObsCreateUrl + "vault=v-name&file=note&content=&append=false&overwrite=false"},
		{"Given content", "default-v", "note", "content", false, false, actions.ObsCreateUrl + "vault=default-v&file=note&content=content&append=false&overwrite=false"},
		{"Given content to overwrite", "default-v", "note", "content", false, true, actions.ObsCreateUrl + "vault=default-v&file=note&content=content&append=false&overwrite=true"},
		{"Given content to append", "default-v", "note", "content", true, false, actions.ObsCreateUrl + "vault=default-v&file=note&content=content&append=true&overwrite=false"},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			actions.CreateNote(tt.vaultName, tt.noteName, tt.content, tt.shouldAppend, tt.shouldOverwrite)
			t.Run("Then it should call the uri constructor with the correct parameters", func(t *testing.T) {
				got := actions.CreateNote(tt.vaultName, tt.noteName, tt.content, tt.shouldAppend, tt.shouldOverwrite)
				assert.Equal(t, tt.want, got)
			})
		})
	}
}

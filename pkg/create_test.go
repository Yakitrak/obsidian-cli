package pkg_test

import (
	"github.com/Yakitrak/obsidian-cli/pkg"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateNote(t *testing.T) {
	var calledBaseUri string
	var calledVaultName string
	var calledNoteName string
	var calledContent string
	var calledShouldAppend bool
	var calledShouldOverwrite bool

	mockUriConstructor := func(baseUri string, params map[string]string) string {
		calledBaseUri = baseUri
		calledVaultName = params["vault"]
		calledNoteName = params["file"]
		calledContent = params["content"]
		calledShouldAppend = params["append"] == "true"
		calledShouldOverwrite = params["overwrite"] == "true"
		return ""
	}

	var tests = []struct {
		testName        string
		vaultName       string
		noteName        string
		content         string
		shouldAppend    bool
		shouldOverwrite bool
	}{

		{"Given no content", "v-name", "note", "", false, false},
		{"Given content", "default-v", "note", "content", false, false},
		{"Given content to overwrite", "default-v", "note", "content", false, true},
		{"Given content to append", "default-v", "note", "content", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			pkg.CreateNote(mockUriConstructor, tt.vaultName, tt.noteName, tt.content, tt.shouldAppend, tt.shouldOverwrite)
			t.Run("Then it should call the uri constructor with the correct parameters", func(t *testing.T) {
				assert.Equal(t, pkg.ObsCreateUrl, calledBaseUri)
				assert.Equal(t, tt.vaultName, calledVaultName)
				assert.Equal(t, tt.noteName, calledNoteName)
				assert.Equal(t, tt.content, calledContent)
				assert.Equal(t, tt.shouldAppend, calledShouldAppend)
				assert.Equal(t, tt.shouldOverwrite, calledShouldOverwrite)
			})
		})
	}
}

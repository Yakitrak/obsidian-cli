package utils_test

import (
	"errors"
	"github.com/Yakitrak/obsidian-cli/utils"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUriExecute(t *testing.T) {
	type testCase struct {
		name string
		uri  string
		err  error
	}

	tests := []testCase{
		{
			name: "valid uri",
			uri:  "http://example.com",
			err:  nil,
		},
		{
			name: "invalid uri",
			uri:  "not a uri",
			err:  errors.New("failed to open URI: not a uri"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := utils.UriExecute(tt.uri)
			assert.Equal(t, tt.err, err)
		})
	}
}

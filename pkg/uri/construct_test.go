package uri_test

import (
	"fmt"
	"github.com/Yakitrak/obsidian-cli/pkg/uri"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUriConstruct(t *testing.T) {
	baseUri := "base-uri"
	var tests = []struct {
		testName string
		baseUri  string
		in       map[string]string
		want     string
	}{
		{"Empty map", baseUri, map[string]string{}, baseUri},
		{"One key", baseUri, map[string]string{"key": "value"}, fmt.Sprintf("%s?key=value", baseUri)},
		{"Two keys", baseUri, map[string]string{"key1": "value1", "key2": "value2"}, fmt.Sprintf("%s?key1=value1&key2=value2", baseUri)},
		{"Empty value", baseUri, map[string]string{"key": ""}, baseUri},
		{"Mix of empty and non-empty values", baseUri, map[string]string{"key1": "value1", "key2": ""}, fmt.Sprintf("%s?key1=value1", baseUri)},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			// Act
			got := uri.Construct(baseUri, tt.in)
			// Assert
			assert.Equal(t, tt.want, got)
		})
	}
}

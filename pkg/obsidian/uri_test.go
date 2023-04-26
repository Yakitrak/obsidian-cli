package obsidian_test

import (
	"errors"
	"fmt"
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUriConstruct(t *testing.T) {
	baseUri := "base-uri"
	var tests = []struct {
		testName string
		in       map[string]string
		want     string
	}{
		{"Empty map", map[string]string{}, baseUri},
		{"One key", map[string]string{"key": "value"}, fmt.Sprintf("%s?key=value", baseUri)},
		{"Two keys", map[string]string{"key1": "value1", "key2": "value2"}, fmt.Sprintf("%s?key1=value1&key2=value2", baseUri)},
		{"Empty value", map[string]string{"key": ""}, baseUri},
		{"Mix of empty and non-empty values", map[string]string{"key1": "value1", "key2": ""}, fmt.Sprintf("%s?key1=value1", baseUri)},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			// Act
			uriManager := obsidian.Uri{}
			got := uriManager.Construct(baseUri, test.in)
			// Assert
			assert.Equal(t, test.want, got)
		})
	}
}

func TestUriExecute(t *testing.T) {
	// Temporarily override the Run function
	originalOpenerFunc := obsidian.Run
	defer func() { obsidian.Run = originalOpenerFunc }()

	t.Run("Valid URI", func(t *testing.T) {
		obsidian.Run = func(uri string) error {
			return nil
		}
		// Arrange
		uriManager := obsidian.Uri{}
		// Act
		err := uriManager.Execute("https://example.com")
		// Assert
		assert.Equal(t, nil, err)
	})

	t.Run("Invalid URI", func(t *testing.T) {
		obsidian.Run = func(uri string) error {
			return errors.New("mock error")
		}
		// Arrange
		uriManager := obsidian.Uri{}
		// Act
		err := uriManager.Execute("foo")
		// Assert
		assert.Equal(t, obsidian.ExecuteUriError, err.Error())
	})

}

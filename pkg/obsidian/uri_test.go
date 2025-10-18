package obsidian_test

import (
	"errors"
	"net/url"
	"strings"

	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUriConstruct(t *testing.T) {
	baseUri := "base-uri"
	var tests = []struct {
		testName string
		in       map[string]string
		want     map[string]string
	}{
		{"Empty map", map[string]string{}, nil},
		{"One key", map[string]string{"key": "value"}, map[string]string{"key": "value"}},
		{"Two keys", map[string]string{"key1": "value1", "key2": "value2"}, map[string]string{"key1": "value1", "key2": "value2"}},
		{"Empty value", map[string]string{"key": ""}, nil},
		{"Mix of empty and non-empty values", map[string]string{"key1": "value1", "key2": ""}, map[string]string{"key1": "value1"}},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			// Act
			uriManager := obsidian.Uri{}
			got := uriManager.Construct(baseUri, test.in)
			// Assert
			if test.want == nil {
				assert.Equal(t, baseUri, got)
			} else {
				parts := strings.SplitN(got, "?", 2)
				assert.Equal(t, baseUri, parts[0])

				if len(parts) > 1 {
					parsedParams, _ := url.ParseQuery(parts[1])
					for key, expectedValue := range test.want {
						assert.Equal(t, expectedValue, parsedParams.Get(key))
					}
				}
			}
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

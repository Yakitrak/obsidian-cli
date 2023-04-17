package handler_test

import (
	"errors"
	"fmt"
	"github.com/Yakitrak/obsidian-cli/handler"
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
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			uri := handler.Uri{}
			got := uri.Construct(baseUri, tt.in)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestUriExecute(t *testing.T) {
	originalOpenerFunc := handler.OpenerFunc

	defer func() {
		handler.OpenerFunc = originalOpenerFunc
	}()

	t.Run("valid uri", func(t *testing.T) {
		handler.OpenerFunc = func(uri string) error {
			return nil
		}

		uri := handler.Uri{}
		err := uri.Execute("http://example.com")
		assert.Equal(t, nil, err)
	})

	t.Run("invalid uri", func(t *testing.T) {
		handler.OpenerFunc = func(uri string) error {
			return fmt.Errorf("not a uri %s", uri)
		}

		uri := handler.Uri{}
		err := uri.Execute("http://example.com")
		assert.Equal(t, errors.New("failed to open URI: not a uri http://example.com"), err)
	})

}

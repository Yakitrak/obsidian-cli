package uri_test

import (
	"errors"
	"github.com/Yakitrak/obsidian-cli/pkg/uri"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUriExecute(t *testing.T) {
	originalOpenerFunc := uri.Run
	defer func() { uri.Run = originalOpenerFunc }()

	t.Run("valid uri", func(t *testing.T) {
		// Arrange
		uri.Run = func(uri string) error {
			return nil
		}

		// Act
		err := uri.Execute("http://example.com")
		// Assert
		assert.Equal(t, nil, err)
	})

	t.Run("invalid uri", func(t *testing.T) {
		// Arrange
		uri.Run = func(uri string) error {
			return errors.New("not a uri")
		}
		// Act
		err := uri.Execute("foo")
		// Assert
		assert.Equal(t, errors.New("failed to open URI: not a uri"), err)
	})

}

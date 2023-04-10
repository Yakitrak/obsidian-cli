package utils

import (
	"errors"
	"github.com/skratchdot/open-golang/open"
)

func UriExecute(uri string) error {
	err := open.Run(uri)
	if err != nil {
		return errors.New("failed to open URI: not a uri")

	}
	return nil
}

package utils

import (
	"errors"
	"github.com/skratchdot/open-golang/open"
	"net/url"
)

var RunUri = open.Run

func UriConstruct(baseUri string, params map[string]string) string {
	uri := baseUri
	for key, value := range params {
		if value != "" && value != "false" {
			if uri == baseUri {
				uri += "?" + key + "=" + url.PathEscape(value)
			} else {
				uri += "&" + key + "=" + url.PathEscape(value)
			}
		}
	}
	return uri
}

func UriExecute(uri string) error {
	err := RunUri(uri)
	if err != nil {
		return errors.New("failed to open URI: not a uri")

	}
	return nil
}

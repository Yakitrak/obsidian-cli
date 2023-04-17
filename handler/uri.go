package handler

import (
	"fmt"
	"github.com/skratchdot/open-golang/open"
	"net/url"
)

type Uri struct {
}

func (u *Uri) Construct(baseUri string, params map[string]string) string {
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

func (u *Uri) Execute(uri string) error {
	err := OpenerFunc(uri)
	if err != nil {
		return fmt.Errorf("failed to open URI: %s", err)
	}
	return nil
}

var OpenerFunc = open.Run

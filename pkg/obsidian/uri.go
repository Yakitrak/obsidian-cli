package obsidian

import (
	"errors"
	"github.com/skratchdot/open-golang/open"
	"net/url"
)

type Uri struct {
}

type UriManager interface {
	Construct(baseUri string, params map[string]string) string
	Execute(uri string) error
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

var Run = open.Run

func (u *Uri) Execute(uri string) error {
	//fmt.Println("Opening URI: ", uri)
	err := Run(uri)
	if err != nil {
		return errors.New(ExecuteUriError)

	}
	return nil
}

package uri

import (
	"net/url"
)

func Construct(baseUri string, params map[string]string) string {
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

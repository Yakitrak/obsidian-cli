package utils

import (
	"net/url"
)

func UrlConstructor(params map[string]string) string {
	var uri string
	for key, value := range params {
		if value != "" && value != "false" {
			if uri == "" {
				uri += "?" + key + "=" + url.PathEscape(value)
			} else {
				uri += "&" + key + "=" + url.PathEscape(value)
			}
		}
	}
	return uri
}

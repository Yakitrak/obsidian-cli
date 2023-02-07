package utils

import (
	"fmt"
	"net/url"
	"strings"
)

func ExtractParams(obsUrl string) map[string]string {
	queryString := strings.Split(obsUrl, "?")[1]
	params, _ := url.ParseQuery(queryString)
	extractedParams := make(map[string]string)
	for key, value := range params {
		extractedParams[key] = value[0]
	}
	fmt.Println(extractedParams)
	return extractedParams
}
func MapsEqual(m1, m2 map[string]string) bool {
	if len(m1) != len(m2) {
		return false
	}

	for k, v := range m1 {
		if m2[k] != v {
			return false
		}
	}

	return true
}

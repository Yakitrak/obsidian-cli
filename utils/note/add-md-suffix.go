package note

import "strings"

func AddMdSuffix(str string) string {
	if !strings.HasSuffix(str, ".md") {
		return str + ".md"
	}
	return str
}

package utils

import (
	"github.com/skratchdot/open-golang/open"
)

func UriExecute(uri string) {
	// fmt.Println(uri)
	open.Run(uri)
}

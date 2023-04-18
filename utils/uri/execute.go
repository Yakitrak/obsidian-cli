package uri

import (
	"errors"
	"github.com/skratchdot/open-golang/open"
)

var Run = open.Run

func Execute(uri string) error {
	//fmt.Println("Opening URI: ", uri)
	err := Run(uri)
	if err != nil {
		return errors.New("failed to open URI: not a uri")

	}
	return nil
}

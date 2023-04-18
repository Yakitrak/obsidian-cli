package note

import (
	"fmt"
	"os"
)

func Move(originalPath string, newPath string) error {

	o := AddMdSuffix(originalPath)
	n := AddMdSuffix(newPath)

	err := os.Rename(o, n)

	if err != nil {
		return err
	} else {
		message := fmt.Sprintf(`Moved note 
from %s
to %s`, o, n)
		fmt.Println(message)
	}
	return nil
}

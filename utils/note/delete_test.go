package note

import (
	"fmt"
	"os"
)

func Delete(originalPath string) error {
	o := AddMdSuffix(originalPath)
	err := os.Remove(o)
	if err != nil {
		return err
	} else {
		fmt.Println("Deleted note: ", o)
	}
	return nil
}

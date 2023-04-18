package note

import (
	"errors"
	"fmt"
	"os"
)

func Delete(path string) error {
	note := AddMdSuffix(path)
	err := os.Remove(note)
	if err != nil {
		return errors.New("note does not exist")
	} else {
		fmt.Println("Deleted note: ", note)
	}
	return nil
}

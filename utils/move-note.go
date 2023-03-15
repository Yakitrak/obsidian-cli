package utils

import (
	"fmt"
	"log"
	"os"
)

func MoveNote(originalPath string, newPath string) {

	o := AddMdSuffix(originalPath)
	n := AddMdSuffix(newPath)

	e := os.Rename(o, n)

	if e != nil {
		log.Fatal(e)
	} else {
		message := fmt.Sprintf(`Moved note 
		from %s
		to %s`, o, n)
		fmt.Println(message)

	}

}

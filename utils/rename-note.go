package utils

import (
	"log"
	"os"
)

func RenameNote(originalPath string, newPath string) {

	o := originalPath + "md"
	n := newPath + "md"

	e := os.Rename(o, n)
	if e != nil {
		log.Fatal(e)
	}

}

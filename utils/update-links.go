package utils

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

func UpdateLinksInVault(dirPath string, oldNoteName string, newNoteName string) {

	// TODO change regex pattern to work if using #heading or |aliss
	oldText := "[[" + oldNoteName + "]]"
	newText := "[[" + newNoteName + "]]"

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		content, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		newContent := bytes.ReplaceAll(content, []byte(oldText), []byte(newText))

		err = ioutil.WriteFile(path, newContent, info.Mode())
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print("Update links inside: ", dirPath, " from ", oldNoteName, " to ", newNoteName)

}

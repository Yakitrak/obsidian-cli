package utils

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func UpdateLinksInVault(dirPath string, oldNoteName string, newNoteName string) {

	if strings.Contains(oldNoteName, "/") {
		oldNoteName = filepath.Base(oldNoteName)
	}

	if strings.Contains(newNoteName, "/") {
		newNoteName = filepath.Base(newNoteName)
	}

	// TODO change regex pattern to work if using #heading or |aliss
	// Standard note links
	oldNoteStandardText := "[[" + oldNoteName + "]]"
	newNoteStandardText := "[[" + newNoteName + "]]"

	// Aliased note links
	oldNoteWithAliasText := "[[" + oldNoteName + "|"
	newNoteWithAliasText := "[[" + newNoteName + "|"

	// Note links with headings
	oldNoteWithHeadingText := "[[" + oldNoteName + "#"
	newNoteWithHeadingText := "[[" + newNoteName + "#"

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden directories and non-markdown files
		if shouldSkipDirectoryOrFile(info) {
			return nil
		}

		content, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		newContentWithUpdatedStandardLinks := bytes.ReplaceAll(content, []byte(oldNoteStandardText), []byte(newNoteStandardText))
		newContentWithUpdatedStandardAndAliasLinks := bytes.ReplaceAll(newContentWithUpdatedStandardLinks, []byte(oldNoteWithAliasText), []byte(newNoteWithAliasText))
		newContentWithUpdatedStandardAndAliasAndHeadingLinks := bytes.ReplaceAll(newContentWithUpdatedStandardAndAliasLinks, []byte(oldNoteWithHeadingText), []byte(newNoteWithHeadingText))

		err = ioutil.WriteFile(path, newContentWithUpdatedStandardAndAliasAndHeadingLinks, info.Mode())
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

func shouldSkipDirectoryOrFile(info os.FileInfo) bool {
	isDirectory := info.IsDir()
	isHidden := info.Name()[0] == '.'
	isNonMarkdownFile := filepath.Ext(info.Name()) != ".md"

	if isDirectory || isHidden || isNonMarkdownFile {
		return true
	}

	return false
}

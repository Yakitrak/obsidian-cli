package notes

import (
	"fmt"
	"github.com/Yakitrak/obsidian-cli/utils"
	"os"
)

func MoveNote(originalPath string, newPath string) error {

	o := utils.AddMdSuffix(originalPath)
	n := utils.AddMdSuffix(newPath)

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

package actions

import (
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"strconv"
)

type CreateParams struct {
	NoteName        string
	ShouldAppend    bool
	ShouldOverwrite bool
	Content         string
}

func CreateNote(vaultOp obsidian.VaultOperator, uriManager obsidian.UriManager, params CreateParams) error {
	vaultName, err := vaultOp.DefaultName()
	if err != nil {
		return err
	}

	uri := uriManager.Construct(ObsCreateUrl, map[string]string{
		"vault":     vaultName,
		"append":    strconv.FormatBool(params.ShouldAppend),
		"overwrite": strconv.FormatBool(params.ShouldOverwrite),
		"content":   params.Content,
		"file":      params.NoteName,
	})

	err = uriManager.Execute(uri)
	if err != nil {
		return err
	}
	return nil
}

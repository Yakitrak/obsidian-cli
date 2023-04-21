package actions

import (
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
)

type OpenParams struct {
	NoteName string
}

func OpenNote(vaultOp obsidian.VaultOperator, uriManager obsidian.UriManager, params OpenParams) error {
	vaultName, err := vaultOp.DefaultName()
	if err != nil {
		return err
	}

	uri := uriManager.Construct(ObsOpenUrl, map[string]string{
		"vault": vaultName,
		"file":  params.NoteName,
	})

	err = uriManager.Execute(uri)
	if err != nil {
		return err
	}
	return nil
}

package pkg

func OpenNote(uriConstructor UriConstructorFunc, vaultName string, noteName string) string {
	uri := uriConstructor(ObsOpenUrl, map[string]string{
		"file":  noteName,
		"vault": vaultName,
	})

	return uri
}

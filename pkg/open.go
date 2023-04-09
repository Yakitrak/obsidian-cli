package pkg

func OpenNote(uriConstructor UriConstructorFunc, noteName string, vaultName string) string {
	uri := uriConstructor(ObsOpenUrl, map[string]string{
		"file":  noteName,
		"vault": vaultName,
	})

	return uri
}

package pkg

func SearchNotes(uriConstructor UriConstructorFunc, vaultName string, searchText string) string {
	uri := uriConstructor(ObsSearchUrl, map[string]string{
		"query": searchText,
		"vault": vaultName,
	})

	return uri
}

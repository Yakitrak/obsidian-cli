package actions

type UriConstructorFunc func(baseUri string, params map[string]string) string
type FindVaultPathFromNameFunc func(vaultName string, configFilePath string) (string, error)
type MoveNoteFunc func(currentPath string, newPath string) error
type UpdateLinksInVaultFunc func(vaultPath string, currentNoteName string, newNoteName string)

package actions_test

import (
	"testing"
)

func TestMoveNote(t *testing.T) {
	//var calledBaseUri string
	//var calledVaultName string
	//var mockVaultPath = "vault-path"
	//var calledCurrentPath string
	//var calledNewPath string
	//var calledCurrentNoteName string
	//var calledNewNoteName string
	//
	//mockUriConstructor := func(baseUri string, params map[string]string) string {
	//	calledBaseUri = baseUri
	//	calledVaultName = params["vault"]
	//	return ""
	//}
	//
	//mockFindVaultPathFromName := func(vaultName string, configFilePath string) (string, error) {
	//	return mockVaultPath, nil
	//}
	//
	//mockFindVaultPathFromNameWithError := func(vaultName string, configFilePath string) (string, error) {
	//	return "", errors.New("Cannot locate vault " + vaultName)
	//}
	//
	//mockMoveNote := func(currentPath string, newPath string) error {
	//	calledCurrentPath = currentPath
	//	calledNewPath = newPath
	//	return nil
	//}
	//
	//mockMoveNoteWithError := func(currentPath string, newPath string) error {
	//	return errors.New("Cannot move note")
	//}
	//
	//mockUpdateLinksInVault := func(vaultPath string, currentNoteName string, newNoteName string) {
	//	calledCurrentNoteName = currentNoteName
	//	calledNewNoteName = newNoteName
	//	return
	//}
	//
	//var tests = []struct {
	//	testName        string
	//	vaultName       string
	//	currentNoteName string
	//	newNoteName     string
	//}{
	//
	//	{"Given vault", "v-name", "current-note-name", "new-note-name"},
	//}
	//for _, tt := range tests {
	//	t.Run(tt.testName, func(t *testing.T) {
	//
	//		pkg.MoveNote(mockUriConstructor, mockFindVaultPathFromName, mockMoveNote, mockUpdateLinksInVault, tt.vaultName, tt.currentNoteName, tt.newNoteName)
	//
	//		t.Run("mockUriConstructor call", func(t *testing.T) {
	//			assert.Equal(t, calledBaseUri, pkg.obsOpenUrl, "unexpected base URI")
	//			assert.Equal(t, calledVaultName, tt.vaultName, "unexpected vault name")
	//
	//		})
	//
	//		t.Run("mockMoveNote call", func(t *testing.T) {
	//			assert.Equal(t, calledCurrentPath, "vault-path/"+tt.currentNoteName, "unexpected current path")
	//			assert.Equal(t, calledNewPath, "vault-path/"+tt.newNoteName, "unexpected new path")
	//		})
	//
	//		t.Run("mockUpdateLinksInVault call", func(t *testing.T) {
	//			assert.Equal(t, calledCurrentNoteName, tt.currentNoteName, "unexpected current note name")
	//			assert.Equal(t, calledNewNoteName, tt.newNoteName, "unexpected new note name")
	//		})
	//	})
	//
	//	t.Run("Given vault with error", func(t *testing.T) {
	//		_, err := pkg.MoveNote(mockUriConstructor, mockFindVaultPathFromNameWithError, mockMoveNote, mockUpdateLinksInVault, tt.vaultName, tt.currentNoteName, tt.newNoteName)
	//		fmt.Println(err)
	//		assert.Equal(t, err.Error(), "Cannot locate vault "+tt.vaultName, "unexpected error")
	//	})
	//
	//	t.Run("Given error with MoveNote", func(t *testing.T) {
	//		_, err := pkg.MoveNote(mockUriConstructor, mockFindVaultPathFromName, mockMoveNoteWithError, mockUpdateLinksInVault, tt.vaultName, tt.currentNoteName, tt.newNoteName)
	//		fmt.Println(err)
	//		assert.Equal(t, err.Error(), "Cannot move note", "unexpected error")
	//	})
	//
	//}

}

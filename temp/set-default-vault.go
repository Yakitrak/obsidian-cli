//package temp
//
//import (
//	"encoding/json"
//	"fmt"
//	"github.com/Yakitrak/obsidian-cli/utils"
//	"os"
//)
//
//type Config struct {
//	DefaultVaultName string `json:"default_vault_name"` // TODO make lower case?
//}
//
//func SetDefaultVault(name string, obsConfigPath string) error {
//	jsonContent, err := json.Marshal(Config{DefaultVaultName: name})
//	if err != nil {
//		return fmt.Errorf("failed to save default vault to configuration: %s", err)
//	}
//
//	// create config dir
//	err = os.MkdirAll(obsConfigPath, os.ModePerm)
//	if err != nil {
//		return fmt.Errorf("failed to create default vault directory %s", err)
//	}
//
//	// create file
//	configFile, err := os.Create(obsConfigPath + utils.ObsConfigName)
//
//	if err != nil {
//		return fmt.Errorf("failed to create default vault configuration %s", err)
//	}
//
//	// write file
//	_, err = configFile.WriteString(string(jsonContent))
//	if err != nil {
//		return fmt.Errorf("failed to write default vault to configuration %s", err)
//
//	}
//	return nil
//}
//
////if err != nil {
////	return "(, fmt.Errorf"VaultInformation name has unexpected character", err)
////}
////
////// Get default user config dir
////dirname, err := os.UserConfigDir()
////if err != nil {
////	return "(, fmt.Errorf"User config directory not found", err)
////}

package utils_test

import (
	"encoding/json"
	"github.com/Yakitrak/obsidian-cli/utils"
	"log"
	"os"
	"testing"
)

func TestGetDefaultVaultWithVaultName(t *testing.T) {
	expectedVaultName := "myVault"

	// Call the function with a non-empty vaultName
	result := utils.GetDefaultVault(expectedVaultName)

	// Assert that the result matches the expected vaultName
	if result != expectedVaultName {
		t.Errorf("Expected vault name '%s', but got '%s'", expectedVaultName, result)
	}
}

func TestGetDefaultVaultWithoutVaultName(t *testing.T) {
	expectedVaultName := "defaultVault"

	// Create a temporary config file with the default vault name
	tmpfile, err := os.CreateTemp("", "config.json")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	config := utils.Config{
		DefaultVaultName: expectedVaultName,
	}
	configBytes, err := json.Marshal(config)
	if err != nil {
		log.Fatal(err)
	}
	_, err = tmpfile.Write(configBytes)
	if err != nil {
		log.Fatal(err)
	}
	tmpfile.Close()

	// Call the function without a vaultName
	result := utils.GetDefaultVault("")

	// Assert that the result matches the expected default vault name
	if result != expectedVaultName {
		t.Errorf("Expected default vault name '%s', but got '%s'", expectedVaultName, result)
	}
}

//func TestGetDefaultVaultWithInvalidConfigFile(t *testing.T) {

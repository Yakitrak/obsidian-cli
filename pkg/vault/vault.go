package vault

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Yakitrak/obsidian-cli/pkg/config"
	"github.com/Yakitrak/obsidian-cli/pkg/note"
	"os"
	"path/filepath"
	"strings"
)

type Vault struct {
	Name string
}

type CliConfig struct {
	DefaultVaultName string `json:"default_vault_name"`
}

type ObsidianVaultConfig struct {
	Vaults map[string]struct {
		Path string `json:"path"`
	} `json:"vaults"`
}

var CliConfigPath = config.CliPath
var ObsidianConfigFile = config.ObsidianFile
var JsonMarshal = json.Marshal

func (v *Vault) SetDefaultName(name string) error {
	// marshal vault name to json
	cliConfig := CliConfig{DefaultVaultName: name}
	jsonContent, err := JsonMarshal(cliConfig)
	if err != nil {
		return fmt.Errorf("failed to save default vault to json: %s", err)
	}

	// get cliConfig path
	obsConfigDir, obsConfigFile, err := CliConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get user cliConfig directory %s", err)
	}
	// create directory
	err = os.MkdirAll(obsConfigDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create default vault directory %s", err)
	}

	// create and write file
	err = os.WriteFile(obsConfigFile, jsonContent, 0644)
	if err != nil {
		return fmt.Errorf("failed to create default vault configuration file %s", err)
	}

	v.Name = name

	return nil
}

func (v *Vault) DefaultName() (string, error) {
	if v.Name != "" {
		return v.Name, nil
	}

	// get cliConfig paths
	_, cliConfigFile, err := CliConfigPath()
	if err != nil {
		return "", fmt.Errorf("failed to get user cliConfig directory %s", err)
	}

	// read cliConfig
	content, err := os.ReadFile(cliConfigFile)
	if err != nil {
		return "", fmt.Errorf("cannot find vault cliConfig. please use set-default command to set default vault or use --vault: %s", err)
	}

	// retrieve value
	cliConfig := CliConfig{}
	err = json.Unmarshal(content, &cliConfig)

	if err != nil {
		return "", fmt.Errorf("could not retrieve default vault %s", err)
	}

	if cliConfig.DefaultVaultName == "" {
		return "", fmt.Errorf("could not read value of default vault %s", err)
	}

	v.Name = cliConfig.DefaultVaultName
	return cliConfig.DefaultVaultName, nil
}

func (v *Vault) Path() (string, error) {
	obsidianConfigFile, err := ObsidianConfigFile()
	if err != nil {
		return "", fmt.Errorf("failed to get obsidian config file %s", err)
	}

	content, err := os.ReadFile(obsidianConfigFile)

	if err != nil {
		return "", fmt.Errorf("obsidian config file cannot be found: %s", err)
	}

	vaultsContent := ObsidianVaultConfig{}
	err = json.Unmarshal(content, &vaultsContent)

	if err != nil {
		return "", fmt.Errorf("obsidian config file cannot be parsed: %s", err)
	}

	for _, element := range vaultsContent.Vaults {
		if strings.HasSuffix(element.Path, v.Name) {
			return element.Path, nil
		}
	}

	return "", errors.New("Vault cannot be found. Please ensure the vault is set up on Obsidian.")
}

func (v *Vault) UpdateNoteLinks(vaultPath string, oldNoteName string, newNoteName string) error {
	err := filepath.Walk(vaultPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if note.ShouldSkipDirectoryOrFile(info) {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		oldNoteLinks := note.GenerateNoteLinkTexts(oldNoteName)
		newNoteLinks := note.GenerateNoteLinkTexts(newNoteName)

		content = note.ReplaceContent(content, map[string]string{
			oldNoteLinks[0]: newNoteLinks[0],
			oldNoteLinks[1]: newNoteLinks[1],
			oldNoteLinks[2]: newNoteLinks[2],
		})

		err = os.WriteFile(path, content, info.Mode())
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

var errUnknownDomain = errors.New("unknown domain")

type configData struct {
	DefaultDomain string
	DomainTokens  map[string]string
}

// getConfigFilePath returns the path of a given file within the UserConfigDir.
func getConfigFilePath(filename string) (string, error) {
	configApplicationName := "slack-status-cli"

	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("error getting current config: %w", err)
	}

	configDir = filepath.Join(configDir, configApplicationName)
	_ = os.MkdirAll(configDir, 0o755)
	configFile := filepath.Join(configDir, filename)

	// Handle migration of old config file path
	// NOTE: Will be removed in future versions
	if !fileExists(configFile) {
		// Get old config path to see if we should migrate
		userHomeDir, _ := os.UserHomeDir()
		legacyConfigFile := filepath.Join(
			userHomeDir,
			".config",
			configApplicationName,
			filename,
		)

		if fileExists(legacyConfigFile) {
			log.Printf("Migrating config from %s to %s\n", legacyConfigFile, configFile)

			err = os.Rename(legacyConfigFile, configFile)
			if err != nil {
				err = fmt.Errorf(
					"error migrating old config from %s: %w",
					legacyConfigFile,
					err,
				)
			}
		}
	}

	return configFile, err
}

// readConfig returns the current configuration
func readConfig() (*configData, error) {
	configPath, err := getConfigFilePath("config.json")
	if err != nil {
		return nil, err
	}

	if !fileExists(configPath) {
		return &configData{DomainTokens: map[string]string{}}, nil
	}

	content, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading config from file: %w", err)
	}

	var config configData

	err = json.Unmarshal(content, &config)
	if err != nil {
		return nil, fmt.Errorf("failed parsing json from config file: %w", err)
	}

	return &config, nil
}

// writeConfig writes the provided config data
func writeConfig(config configData) error {
	configPath, err := getConfigFilePath("config.json")
	if err != nil {
		return err
	}

	contents, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed converting config to json: %w", err)
	}

	if err = ioutil.WriteFile(configPath, contents, 0o600); err != nil {
		return fmt.Errorf("error writing config to file: %w", err)
	}

	return nil
}

// getDefaultLogin returns the default login from the configuration
func getDefaultLogin() (string, error) {
	config, err := readConfig()
	if err != nil {
		return "", err
	}

	accessToken, exists := config.DomainTokens[config.DefaultDomain]
	if !exists {
		return "", errUnknownDomain
	}

	return accessToken, nil
}

// getLogin returns the token for a specified login domain
func getLogin(domain string) (string, error) {
	config, err := readConfig()
	if err != nil {
		return "", err
	}

	accessToken, exists := config.DomainTokens[domain]
	if !exists {
		return "", errUnknownDomain
	}

	return accessToken, nil
}

// saveLogin writes the provided token to the provided domain
func saveLogin(domain, accessToken string) error {
	config, err := readConfig()
	if err != nil {
		return err
	}

	config.DomainTokens[domain] = accessToken

	// If this is the only domain, make it default
	if len(config.DomainTokens) == 1 {
		config.DefaultDomain = domain
	}

	return writeConfig(*config)
}

// saveDefaultLogin saves the specified domain as the default
func saveDefaultLogin(domain string) error {
	config, err := readConfig()
	if err != nil {
		return err
	}

	_, exists := config.DomainTokens[domain]
	if !exists {
		return fmt.Errorf("cannot set domain to default: %w", errUnknownDomain)
	}

	config.DefaultDomain = domain

	return writeConfig(*config)
}

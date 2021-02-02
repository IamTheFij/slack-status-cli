package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

var (
	errUnknownDomain = errors.New("unknown domain")
)

type configData struct {
	DefaultDomain string
	DomainTokens  map[string]string
}

// getConfigFilePath returns the path of a given file within the config folder.
// The config folder will be created in ~/.local/config/slack-status-cli if it does not exist.
func getConfigFilePath(filename string) string {
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		configHome = "~/.local/config"
	}

	configDir := filepath.Join(configHome, "slack-status-cli")
	_ = os.MkdirAll(configDir, 0755)

	return filepath.Join(configDir, filename)
}

// readConfig returns the current configuration
func readConfig() (*configData, error) {
	configPath := getConfigFilePath("config.json")

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
	configPath := getConfigFilePath("config.json")

	contents, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed converting config to json: %w", err)
	}

	if err = ioutil.WriteFile(configPath, contents, 0600); err != nil {
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

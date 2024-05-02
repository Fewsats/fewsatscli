package config

import (
	"fmt"
	"log/slog"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

const (
	// baseURL is the base URL for the Fewsats API.
	baseURL = "https://api.fewsats.com"
)

var (
	loadedConfig  *Config
	defaultConfig = map[string]string{
		"LOG_LEVEL": "info",
	}
)

type Config struct {
	APIKey    string
	Domain    string
	AlbyToken string
	LogLevel  string
	ConfigDir string
}

func getDefaultConfigContent() string {
	var lines []string
	for key, value := range defaultConfig {
		lines = append(lines, fmt.Sprintf(`%s="%s"`, key, value))
	}
	return strings.Join(lines, "\n")
}

func GetConfig() (*Config, error) {
	if loadedConfig != nil {
		return loadedConfig, nil
	}

	// Get the current user
	usr, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("unable to get current OS user: %w", err)
	}

	// Construct the path to the config file (~/.fewsats/config)
	configDir := filepath.Join(usr.HomeDir, ".fewsats")
	configFilePath := filepath.Join(configDir, "config")

	// Check if the config file exists
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		slog.Debug("Config file not found. Creating default config file...\n")
		defaultContent := getDefaultConfigContent()
		// Create the file with default content if it does not exist
		if err := os.WriteFile(configFilePath, []byte(defaultContent), 0644); err != nil {
			return nil, fmt.Errorf("failed to create default config file: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("failed to check config file: %w", err)
	}

	// Load the config file
	err = godotenv.Load(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("unable to load config file: %w", err)
	}

	// Set global config from file
	domain, exists := os.LookupEnv("DOMAIN")
	if !exists {
		domain = baseURL
	}

	apiKey := os.Getenv("APIKEY")
	albyToken := os.Getenv("ALBY_TOKEN")
	logLevel, exists := os.LookupEnv("LOG_LEVEL")
	if !exists {
		logLevel = "info"
	}

	loadedConfig = &Config{
		APIKey:    apiKey,
		Domain:    domain,
		AlbyToken: albyToken,
		LogLevel:  logLevel,
		ConfigDir: configDir,
	}

	return loadedConfig, nil
}

package config

import (
	"fmt"
	"log/slog"
	"os"
	"os/user"
	"path/filepath"

	"github.com/joho/godotenv"
)

const (
	// baseURL is the base URL for the Fewsats API.
	baseURL = "https://api.fewsats.com"
)

var (
	loadedConfig *Config
)

type Config struct {
	APIKey    string
	Domain    string
	AlbyToken string
	LogLevel  string
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

	// Construct the path to the .fewsatscli file
	filepath := filepath.Join(usr.HomeDir, ".fewsatscli")

	// Check if the file exists
	if _, err := os.Stat(filepath); err == nil {
		slog.Debug("Loading .fewsatscli file...\n")

		// If the file exists, load it
		err = godotenv.Load(filepath)
		if err != nil {
			return nil, fmt.Errorf("unable to load .fewsatscli file: %w", err)
		}
	}

	// Set global config from .fewsatscli file
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
	}

	return loadedConfig, nil
}

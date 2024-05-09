package config

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"gopkg.in/ini.v1"
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
	Domain     string
	AlbyToken  string
	LogLevel   string
	ConfigDir  string
	DBFilePath string
}

func getConfigSection(configFilePath, profile string) (*ini.Section, error) {
	// Load or create the config file
	cfg, err := ini.LoadSources(ini.LoadOptions{}, configFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			// File does not exist, create it
			cfg = ini.Empty()
		} else {
			return nil, fmt.Errorf("failed to load config file: %w", err)
		}
	}

	// Check if the profile section exists
	section, err := cfg.GetSection(profile)
	if err != nil {
		if section, err = cfg.NewSection(profile); err != nil {
			return nil, fmt.Errorf("failed to create profile section: %w", err)
		}
		// Populate default settings for new profile
		for key, value := range defaultConfig {
			section.NewKey(key, value)
		}
		// Save the new profile section to file
		if err = cfg.SaveTo(configFilePath); err != nil {
			return nil, fmt.Errorf("failed to save new profile to config file: %w", err)
		}
	}

	return section, nil
}

func GetConfig() (*Config, error) {
	if loadedConfig != nil {
		return loadedConfig, nil
	}

	profile := os.Getenv("PROFILE")

	// Get the current user
	usr, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("unable to get current OS user: %w", err)
	}

	// Construct the path to
	// * the config file (~/.fewsats/config)
	// * the db file (~/.fewsats/{profile}.db)
	configDir := filepath.Join(usr.HomeDir, ".fewsats")
	configFilePath := filepath.Join(configDir, "config")
	dbFilePath := filepath.Join(configDir, fmt.Sprintf("%s.db", profile))

	section, err := getConfigSection(configFilePath, profile)
	if err != nil {
		return nil, fmt.Errorf("failed to get profile config: %w", err)
	}

	domain := section.Key("DOMAIN").MustString(baseURL)
	albyToken := section.Key("ALBY_TOKEN").MustString("")
	logLevel := section.Key("LOG_LEVEL").MustString("info")

	loadedConfig = &Config{
		Domain:     domain,
		AlbyToken:  albyToken,
		LogLevel:   logLevel,
		ConfigDir:  configDir,
		DBFilePath: dbFilePath,
	}

	return loadedConfig, nil
}

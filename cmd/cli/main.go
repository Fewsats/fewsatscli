package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/fewsats/fewsatscli/account"
	"github.com/fewsats/fewsatscli/apikeys"
	"github.com/fewsats/fewsatscli/config"
	"github.com/fewsats/fewsatscli/storage"
	"github.com/fewsats/fewsatscli/store"
	"github.com/fewsats/fewsatscli/version"
	"github.com/urfave/cli/v2"
)

func main() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Failed to get user home directory:", err)
	}

	fewsatsDir := filepath.Join(homeDir, ".fewsats")
	if _, err := os.Stat(fewsatsDir); os.IsNotExist(err) {
		err = os.Mkdir(fewsatsDir, 0755)
		if err != nil {
			log.Fatal("Failed to create .fewsats directory:", err)
		}
	}

	app := &cli.App{
		Name:                 "Fewsats CLI",
		Usage:                "Interact with the Fewsats Platform.",
		Version:              version.Version(),
		EnableBashCompletion: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "profile",
				Value: "default",
				Usage: "Specify the configuration profile",
			},
		},
		Before: func(c *cli.Context) error {
			os.Setenv("PROFILE", c.String("profile"))
			cfg, err := config.GetConfig()
			if err != nil {
				return nil
			}

			store, err := store.NewStore(cfg.DBFilePath)
			if err != nil {
				log.Fatal("Failed to create store:", err)
			}

			err = store.InitSchema()
			if err != nil {
				log.Fatal("Failed to initialize database schema:", err)
			}

			// Set slog level to debug.
			switch cfg.LogLevel {
			case "info":
				slog.SetLogLoggerLevel(slog.LevelInfo)
			case "debug":
				slog.SetLogLoggerLevel(slog.LevelDebug)
			case "warn":
				slog.SetLogLoggerLevel(slog.LevelWarn)
			case "error":
				slog.SetLogLoggerLevel(slog.LevelError)
			}

			return nil
		},
		Commands: []*cli.Command{
			account.Command(),
			apikeys.Command(),
			storage.Command(),
		},
	}

	err = app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
	}
}

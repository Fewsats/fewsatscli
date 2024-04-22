package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/fewsats/fewsatscli/account"
	"github.com/fewsats/fewsatscli/apikeys"
	"github.com/fewsats/fewsatscli/config"
	"github.com/fewsats/fewsatscli/storage"
	"github.com/fewsats/fewsatscli/version"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:                 "Fewsats CLI",
		Usage:                "Interact with the Fewsats Platform.",
		Version:              version.Version(),
		EnableBashCompletion: true,
		Before: func(c *cli.Context) error {
			cfg, err := config.GetConfig()
			if err != nil {
				return nil
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

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
	}
}

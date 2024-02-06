package main

import (
	"fmt"
	"os"

	"github.com/fewsats/fewsatscli/recipes"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:                 "Fewsats CLI",
		Usage:                "Interact with the Fewsats API.",
		EnableBashCompletion: true,
		Commands: []*cli.Command{
			recipes.Command(),
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
	}
}

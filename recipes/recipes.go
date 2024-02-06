package recipes

import (
	"github.com/urfave/cli/v2"
)

const (
	baseURL         = "https://api.fewsats.com/v0"
	albyURL         = "https://api.getalby.com"
	contentTypeJson = "application/json"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:  "recipes",
		Usage: "Interact with code recipes.",
		Subcommands: []*cli.Command{
			AddCodeRecipeCommand,
			GetCodeRecipeCommand,
			ExecuteCodeRecipeCommand,
		},
	}
}

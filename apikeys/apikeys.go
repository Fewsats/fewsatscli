package apikeys

import (
	"github.com/urfave/cli/v2"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:  "apikeys",
		Usage: "Interact with api keys.",
		Subcommands: []*cli.Command{
			createCommand,
			listCommand,
			disableCommand,
		},
	}
}

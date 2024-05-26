package macaroons

import (
	"github.com/urfave/cli/v2"
)

// Command creates the macaroons command with subcommands.
func Command() *cli.Command {
	return &cli.Command{
		Name:  "macaroons",
		Usage: "Interact with macaroon tokens.",
		Subcommands: []*cli.Command{
			decodeCommand,
		},
	}
}

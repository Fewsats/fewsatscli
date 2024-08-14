package gateway

import (
	"github.com/urfave/cli/v2"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:  "gateway",
		Usage: "Interact with gateways.",
		Subcommands: []*cli.Command{
			createCommand,
		},
	}
}

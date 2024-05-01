package account

import (
	"github.com/urfave/cli/v2"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:  "account",
		Usage: "Interact with your account.",
		Subcommands: []*cli.Command{
			signUpCommand,
			loginCommand,
		},
	}
}

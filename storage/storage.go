package storage

import (
	"github.com/urfave/cli/v2"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:  "storage",
		Usage: "Interact with storage services.",
		Subcommands: []*cli.Command{
			uploadFileCommand,
			downloadFileCommand,
		},
	}
}

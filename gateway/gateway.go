package gateway

import (
	"time"

	"github.com/urfave/cli/v2"
)

type Gateway struct {
	ExternalID  string    `json:"external_id"`
	Status      string    `json:"status"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Amount      uint64    `json:"amount"`
	Duration    string    `json:"duration"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func Command() *cli.Command {
	return &cli.Command{
		Name:  "gateway",
		Usage: "Interact with gateways.",
		Subcommands: []*cli.Command{
			createCommand,
			deleteCommand,
			listCommand,
			searchCommand,
			accessCommand,
		},
	}
}

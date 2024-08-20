package gateway

import (
	"time"

	"github.com/urfave/cli/v2"
)

type Gateway struct {
	ExternalID   string    `json:"external_id"`
	Status       string    `json:"status"`
	Name         string    `json:"name"`
	TargetURL    string    `json:"target_url"`
	Description  string    `json:"description"`
	PriceInCents uint64    `json:"price_in_cents"`
	Duration     string    `json:"duration"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func Command() *cli.Command {
	return &cli.Command{
		Name:  "gateway",
		Usage: "Interact with gateways.",
		Subcommands: []*cli.Command{
			createCommand,
			getCommand,
			deleteCommand,
			listCommand,
			searchCommand,
			accessCommand,
			updateCommand,
		},
	}
}

package payout

import (
	"github.com/urfave/cli/v2"
)

type Payout struct {
	ID          uint64 `json:"id"`
	TotalAmount uint64 `json:"total_amount"`
	Currency    string `json:"currency"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
}

func Command() *cli.Command {
	return &cli.Command{
		Name:  "payout",
		Usage: "Interact with payouts.",
		Subcommands: []*cli.Command{
			listCommand,
		},
	}
}

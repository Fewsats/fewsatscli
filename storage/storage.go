package storage

import (
	"time"

	"github.com/urfave/cli/v2"
)

type File struct {
	ExternalID      string    `json:"external_id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	L402URL         string    `json:"l402_url"`
	Size            uint64    `json:"size"`
	Extension       string    `json:"extension"`
	MimeType        string    `json:"mime_type"`
	CoverURL        string    `json:"cover_url"`
	PriceInUsdCents uint64    `json:"price_in_cents"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	Tags            []string  `json:"tags"`
	Status          string    `json:"status"`
}

func Command() *cli.Command {
	return &cli.Command{
		Name:  "storage",
		Usage: "Interact with storage services.",
		Subcommands: []*cli.Command{
			uploadFileCommand,
			downloadFileCommand,
			listCommand,
			getCommand,
			updateCommand,
			deleteCommand,
			searchCommand,
		},
	}
}

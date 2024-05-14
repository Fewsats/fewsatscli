package storage

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/fewsats/fewsatscli/client"
	"github.com/urfave/cli/v2"
)

type File struct {
	ExternalID      string    `json:"external_id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	Size            uint64    `json:"size"`
	Extension       string    `json:"extension"`
	MimeType        string    `json:"mime_type"`
	StorageURL      string    `json:"storage_url"`
	PriceInUsdCents uint64    `json:"price_in_usd_cents"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

var listCommand = &cli.Command{
	Name:   "list",
	Usage:  "List all files.",
	Action: listFiles,
}

func listFiles(c *cli.Context) error {
	err := client.RequiresLogin()
	if err != nil {
		slog.Debug("Failed to check if user is logged in.", "error", err)
		return cli.Exit("You need to log in to run this command.", 1)
	}

	client, err := client.NewHTTPClient()
	if err != nil {
		slog.Debug("Failed to create HTTP client.", "error", err)
		return cli.Exit("Failed to create HTTP client.", 1)
	}

	resp, err := client.ExecuteRequest(http.MethodGet, "/v0/storage?limit=100", nil)
	if err != nil {
		slog.Debug("Failed to execute request.", "error", err)
		return cli.Exit("Failed to execute request.", 1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return cli.Exit(fmt.Sprintf("Failed request status code: %d", resp.StatusCode), 1)
	}

	var response struct {
		Files []File `json:"files"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return cli.Exit("Failed to decode files.", 1)
	}

	jsonOutput, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return cli.Exit("Failed to marshal JSON.", 1)
	}

	fmt.Println(string(jsonOutput))

	return nil
}

package storage

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"text/tabwriter"
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

func printFiles(files []File) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.Debug)
	fmt.Fprintln(w, "Name\t Updated At\t URL")
	for _, file := range files {
		updatedAt := file.UpdatedAt.Local().Format("2006-01-02 15:04")
		fmt.Fprintf(w, "%s\t %s\t %s\n", file.Name, updatedAt, file.StorageURL)
	}
	w.Flush()
}

func listFiles(c *cli.Context) error {
	err := client.RequiresLogin()
	if err != nil {
		return cli.Exit("You need to log in to run this command.", 1)
	}

	client, err := client.NewHTTPClient()
	if err != nil {
		slog.Debug("Failed to create HTTP client.", "error", err)
		return cli.Exit("Failed to create HTTP client.", 1)
	}

	resp, err := client.ExecuteRequest(http.MethodGet, "/v0/storage/files", nil)
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

	printFiles(response.Files)
	return nil
}

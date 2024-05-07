package storage

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/fewsats/fewsatscli/client"
	"github.com/urfave/cli/v2"
)

const (
	// searchFilesPath is the path to the search endpoint.
	searchFilesPath = "/v0/storage/search"
)

var searchCommand = &cli.Command{
	Name:   "search",
	Usage:  "Search files.",
	Action: searchFiles,
}

func searchFiles(c *cli.Context) error {
	client, err := client.NewHTTPClient()
	if err != nil {
		slog.Debug("Failed to create HTTP client.", "error", err)
		return cli.Exit("Failed to create HTTP client.", 1)
	}

	resp, err := client.ExecuteRequest(http.MethodGet, searchFilesPath, nil)
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
		return cli.Exit("Failed to decode response.", 1)
	}

	// Marshal the response struct to JSON
	jsonOutput, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return cli.Exit("Failed to marshal JSON.", 1)
	}

	fmt.Println(string(jsonOutput))

	return nil
}

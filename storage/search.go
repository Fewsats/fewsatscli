package storage

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"

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
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "name",
			Usage: "Search query to filter files by name",
		},
	},
}

func searchFiles(c *cli.Context) error {
	client, err := client.NewHTTPClient()
	if err != nil {
		slog.Debug("Failed to create HTTP client.", "error", err)
		return cli.Exit("Failed to create HTTP client.", 1)
	}

	requestURL := fmt.Sprintf("%s?limit=100", searchFilesPath)

	// Retrieve the query parameter from the command line, if provided
	nameQuery := c.String("name")
	if nameQuery != "" {
		requestURL += fmt.Sprintf("&name=%s", url.QueryEscape(nameQuery))
	}

	resp, err := client.ExecuteRequest(http.MethodGet, requestURL, nil)
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

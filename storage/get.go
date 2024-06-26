package storage

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/fewsats/fewsatscli/client"
	"github.com/urfave/cli/v2"
)

var getCommand = &cli.Command{
	Name:      "get",
	Usage:     "Get a file by ID.",
	ArgsUsage: "<file_id>",
	Action:    getFile,
}

func getFile(c *cli.Context) error {
	if c.Args().Len() < 1 {
		return cli.Exit("missing <file_id> argument", 1)
	}

	fileID := c.Args().Get(0)

	client, err := client.NewHTTPClient()
	if err != nil {
		slog.Debug("Failed to create HTTP client.", "error", err)
		return cli.Exit("Failed to create HTTP client.", 1)
	}

	resp, err := client.ExecuteRequest(http.MethodGet, fmt.Sprintf("/v0/storage/%s", fileID), nil)
	if err != nil {
		slog.Debug("Failed to execute request.", "error", err)
		return cli.Exit("Failed to execute request.", 1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return cli.Exit(fmt.Sprintf("Failed request status code: %d", resp.StatusCode), 1)
	}

	var response struct {
		File File `json:"file"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return cli.Exit("Failed to decode file.", 1)
	}
	// Marshal the response struct to JSON
	jsonOutput, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return cli.Exit("Failed to marshal JSON.", 1)
	}

	fmt.Println(string(jsonOutput))

	return nil
}

package storage

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/fewsats/fewsatscli/client"
	"github.com/urfave/cli/v2"
)

var deleteCommand = &cli.Command{
	Name:      "delete",
	Usage:     "Delete a file by its external ID.",
	ArgsUsage: "<external_id>",
	Action:    deleteFile,
}

func deleteFile(c *cli.Context) error {
	err := client.RequiresLogin()
	if err != nil {
		return cli.Exit("You need to log in to run this command.", 1)
	}

	if c.Args().Len() < 1 {
		return cli.Exit("missing <external_id> argument", 1)
	}

	externalID := c.Args().Get(0)

	client, err := client.NewHTTPClient()
	if err != nil {
		slog.Debug("Failed to create HTTP client.", "error", err)
		return cli.Exit("Failed to create HTTP client.", 1)
	}

	resp, err := client.ExecuteRequest(http.MethodDelete, fmt.Sprintf("/v0/storage/%s", externalID), nil)
	if err != nil {
		slog.Debug("Failed to execute request.", "error", err)
		return cli.Exit("Failed to execute request.", 1)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return cli.Exit(fmt.Sprintf("Failed request status code: %d", resp.StatusCode), 1)
	}

	fmt.Println("File deleted successfully.")
	return nil
}

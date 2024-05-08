package storage

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/fewsats/fewsatscli/client"
	"github.com/urfave/cli/v2"
)

var updateCommand = &cli.Command{
	Name:      "update",
	Usage:     "Update a file by ID.",
	ArgsUsage: "<file_id>",
	Action:    updateFile,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "name",
			Usage: "Updated name of the file",
		},
		&cli.StringFlag{
			Name:  "description",
			Usage: "Updated description of the file",
		},
		&cli.Uint64Flag{
			Name:  "price",
			Usage: "Updated price of the file in USD cents",
		},
	},
}

func updateFile(c *cli.Context) error {
	err := client.RequiresLogin()
	if err != nil {
		return cli.Exit("You need to log in to run this command.", 1)
	}

	if c.Args().Len() < 1 {
		return cli.Exit("missing <file_id> argument", 1)
	}

	if c.NumFlags() == 0 {
		return cli.Exit("at least one flag (--name, --description, or --price) is required", 1)
	}

	fileID := c.Args().Get(0)

	updateData := make(map[string]interface{})
	if name := c.String("name"); name != "" {
		updateData["name"] = name
	}
	if description := c.String("description"); description != "" {
		updateData["description"] = description
	}
	if price := c.Uint64("price"); price > 0 {
		updateData["price_in_usd_cents"] = price
	}

	client, err := client.NewHTTPClient()
	if err != nil {
		slog.Debug("Failed to create HTTP client.", "error", err)
		return cli.Exit("Failed to create HTTP client.", 1)
	}

	updateDataJSON, err := json.Marshal(updateData)
	if err != nil {
		slog.Debug("Failed to marshal update data.", "error", err)
		return cli.Exit("Failed to marshal update data.", 1)
	}

	resp, err := client.ExecuteRequest(http.MethodPatch, fmt.Sprintf("/v0/storage/%s", fileID), updateDataJSON)
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

	fmt.Println("File updated successfully.")

	return nil
}

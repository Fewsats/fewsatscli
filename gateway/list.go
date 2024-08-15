package gateway

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/fewsats/fewsatscli/client"
	"github.com/urfave/cli/v2"
)

var listCommand = &cli.Command{
	Name:   "list",
	Usage:  "List all gateways.",
	Action: listGateways,
	Flags: []cli.Flag{
		&cli.IntFlag{
			Name:  "limit",
			Usage: "Limit the number of results",
			Value: 100,
		},
		&cli.IntFlag{
			Name:  "offset",
			Usage: "Offset for pagination",
			Value: 0,
		},
	},
}

func listGateways(c *cli.Context) error {
	err := client.RequiresLogin()
	if err != nil {
		return cli.Exit("You need to log in to run this command.", 1)
	}

	client, err := client.NewHTTPClient()
	if err != nil {
		slog.Debug("Failed to create HTTP client.", "error", err)
		return cli.Exit("Failed to create HTTP client.", 1)
	}

	limit := c.Int("limit")
	offset := c.Int("offset")

	url := fmt.Sprintf("/v0/gateway?limit=%d&offset=%d", limit, offset)
	resp, err := client.ExecuteRequest(http.MethodGet, url, nil)
	if err != nil {
		slog.Debug("Failed to execute request.", "error", err)
		return cli.Exit("Failed to execute request.", 1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return cli.Exit(fmt.Sprintf("Failed request status code: %d", resp.StatusCode), 1)
	}

	var response struct {
		Gateways []Gateway `json:"gateways"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return cli.Exit("Failed to decode gateways.", 1)
	}

	jsonOutput, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return cli.Exit("Failed to marshal JSON.", 1)
	}

	fmt.Println(string(jsonOutput))

	return nil
}

package gateway

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/fewsats/fewsatscli/client"
	"github.com/urfave/cli/v2"
)

var deleteCommand = &cli.Command{
	Name:      "delete",
	Usage:     "Delete a gateway by its ID.",
	ArgsUsage: "<gateway_id>",
	Action:    deleteGateway,
}

func deleteGateway(c *cli.Context) error {
	err := client.RequiresLogin()
	if err != nil {
		return cli.Exit("You need to log in to run this command.", 1)
	}

	if c.Args().Len() < 1 {
		return cli.Exit("missing <gateway_id> argument", 1)
	}

	gatewayID := c.Args().Get(0)

	client, err := client.NewHTTPClient()
	if err != nil {
		slog.Debug("Failed to create HTTP client.", "error", err)
		return cli.Exit("Failed to create HTTP client.", 1)
	}

	resp, err := client.ExecuteRequest(http.MethodDelete, fmt.Sprintf("/v0/gateway/%s", gatewayID), nil)
	if err != nil {
		slog.Debug("Failed to execute request.", "error", err)
		return cli.Exit("Failed to execute request.", 1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return cli.Exit(fmt.Sprintf("Failed to delete gateway. Status code: %d", resp.StatusCode), 1)
	}

	fmt.Println("Gateway deleted successfully.")
	return nil
}

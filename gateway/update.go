package gateway

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/fewsats/fewsatscli/client"
	"github.com/urfave/cli/v2"
)

var updateCommand = &cli.Command{
	Name:  "update",
	Usage: "Update an existing gateway.",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "id",
			Usage:    "The ID of the gateway to update",
			Required: true,
		},
		&cli.StringFlag{
			Name:  "name",
			Usage: "The new name of the gateway",
		},
		&cli.StringFlag{
			Name:  "description",
			Usage: "The new description of the gateway",
		},
		&cli.Uint64Flag{
			Name:  "price",
			Usage: "The new price in cents for the gateway",
		},
	},
	Action: updateGateway,
}

func updateGateway(c *cli.Context) error {
	err := client.RequiresLogin()
	if err != nil {
		return cli.Exit("You need to log in to run this command.", 1)
	}

	id := c.String("id")
	if id == "" {
		return cli.Exit("missing required --id flag", 1)
	}

	req := UpdateGatewayRequest{
		Name:         c.String("name"),
		Description:  c.String("description"),
		PriceInCents: c.Uint64("price"),
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to marshal request: %v", err), 1)
	}

	client, err := client.NewHTTPClient()
	if err != nil {
		slog.Debug("Failed to create HTTP client.", "error", err)
		return cli.Exit("Failed to create HTTP client.", 1)
	}

	resp, err := client.ExecuteRequest(http.MethodPatch, fmt.Sprintf("/v0/gateway/%s", id), jsonData)
	if err != nil {
		slog.Debug("Failed to execute request.", "error", err)
		return cli.Exit("Failed to execute request.", 1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return cli.Exit(fmt.Sprintf("Failed to update gateway. Status code: %d, Body: %s", resp.StatusCode, string(bodyBytes)), 1)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to read response body: %v", err), 1)
	}

	var response struct {
		Message string `json:"message"`
	}
	err = json.Unmarshal(bodyBytes, &response)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to decode response: %v. Raw response: %s", err, string(bodyBytes)), 1)
	}

	fmt.Println(response.Message)

	return nil
}

type UpdateGatewayRequest struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	PriceInCents uint64 `json:"price_in_cents"`
}

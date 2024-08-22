package gateway

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/fewsats/fewsatscli/client"
	"github.com/urfave/cli/v2"
)

var visitsCommand = &cli.Command{
	Name:      "visits",
	Usage:     "Get visit details for a gateway.",
	ArgsUsage: "<gateway_id>",
	Action:    getGatewayVisits,
}

type VisitData struct {
	Date       string `json:"date"`
	VisitCount int    `json:"visit_count"`
}

type GatewayVisitsResponse struct {
	GatewayVisits struct {
		GatewayID     string      `json:"gateway_id"`
		GatewayVisits []VisitData `json:"gateway_visits"`
	} `json:"gateway_visits"`
}

func getGatewayVisits(c *cli.Context) error {
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

	resp, err := client.ExecuteRequest(http.MethodGet, fmt.Sprintf("/v0/gateway/%s/details", gatewayID), nil)
	if err != nil {
		slog.Debug("Failed to execute request.", "error", err)
		return cli.Exit("Failed to execute request.", 1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return cli.Exit(fmt.Sprintf("Failed request status code: %d", resp.StatusCode), 1)
	}

	var response GatewayVisitsResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return cli.Exit("Failed to decode gateway visits.", 1)
	}

	jsonOutput, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return cli.Exit("Failed to marshal JSON.", 1)
	}

	fmt.Println(string(jsonOutput))

	return nil
}

package users

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/fewsats/fewsatscli/client"
	"github.com/urfave/cli/v2"
)

var getBillingCommand = &cli.Command{
	Name:   "get-billing",
	Usage:  "Get billing information",
	Action: getBillingInformation,
}

func getBillingInformation(c *cli.Context) error {
	client, err := client.NewHTTPClient()
	if err != nil {
		slog.Debug("Failed to create HTTP client.", "error", err)
		return cli.Exit("Failed to create HTTP client", 1)
	}

	resp, err := client.ExecuteRequest(http.MethodGet, "/v0/users/billing", nil)
	if err != nil {
		slog.Debug("Failed to get billing information.", "error", err)
		return cli.Exit("Failed to get billing information", 1)
	}
	defer resp.Body.Close()

	var billingInfo BillingInformation
	if err := json.NewDecoder(resp.Body).Decode(&billingInfo); err != nil {
		slog.Debug("Failed to decode billing information.", "error", err)
		return cli.Exit("Failed to decode billing information", 1)
	}

	fmt.Println("Billing Information:", billingInfo)
	return nil
}

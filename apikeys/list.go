package apikeys

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	// Added to support time.Time in APIKey struct
	"github.com/fewsats/fewsatscli/client"
	"github.com/fewsats/fewsatscli/store"
	"github.com/urfave/cli/v2"
)

var listCommand = &cli.Command{
	Name:   "list",
	Usage:  "List all API keys.",
	Action: listAPIKeys,
}

func listAPIKeys(c *cli.Context) error {
	err := client.RequiresLogin()
	if err != nil {
		slog.Debug("Failed to check if user is logged in.", "error", err)
		return cli.Exit("You need to log in to run this command.", 1)
	}

	client, err := client.NewHTTPClient()
	if err != nil {
		slog.Debug("Failed to create HTTP client.", "error", err)
		return cli.Exit("Failed to create HTTP client.", 1)
	}

	// Try using existing API key
	resp, err := client.ExecuteRequest(http.MethodGet, "/v0/auth/apikeys?limit=100", nil)
	if err != nil {
		slog.Debug("Failed to execute request with API key.", "error", err)
		return cli.Exit("Failed to execute request.", 1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return cli.Exit(fmt.Sprintf("Failed request status code: %d", resp.StatusCode), 1)
	}

	var response struct {
		Keys []store.APIKey `json:"keys"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return cli.Exit("Failed to decode API keys.", 1)
	}

	jsonOutput, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return cli.Exit("Failed to marshal JSON.", 1)
	}

	fmt.Println(string(jsonOutput))

	return nil
}

package apikeys

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"text/tabwriter"
	"time" // Added to support time.Time in APIKey struct

	"github.com/fewsats/fewsatscli/client"
	"github.com/fewsats/fewsatscli/config"
	"github.com/urfave/cli/v2"
)

// APIKey represents an API key that can be used to authenticate requests as
// a given user.
type APIKey struct {
	ID        uint64
	HiddenKey string
	ExpiresAt *time.Time
}

var listCommand = &cli.Command{
	Name:   "list",
	Usage:  "List all API keys.",
	Action: listAPIKeys,
}

func printKeys(keys []APIKey) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.Debug)
	fmt.Fprintln(w, "ID\t Key\t ExpiresAt")
	for _, key := range keys {
		fmt.Fprintf(w, "%d\t %s\t %s\n", key.ID, key.HiddenKey, key.ExpiresAt)
	}
	w.Flush()
}

func listAPIKeys(c *cli.Context) error {
	cfg, err := config.GetConfig()
	if err != nil {
		slog.Debug("Failed to get config.", "error", err)
		return cli.Exit("Failed to get config.", 1)
	}

	client, err := client.NewHTTPClient()
	if err != nil {
		slog.Debug("Failed to create HTTP client.", "error", err)
		return cli.Exit("Failed to create HTTP client.", 1)
	}

	// Try using existing API key
	if cfg.APIKey != "" {
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
			Keys       []APIKey `json:"keys"`
			TotalCount int      `json:"total_count"`
		}
		err = json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			return cli.Exit("Failed to decode API keys.", 1)
		}

		printKeys(response.Keys)
		return nil

	}
	return nil

}
